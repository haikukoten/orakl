package aggregator

import (
	"context"
	"encoding/json"
	"sync"

	"time"

	"bisonai.com/orakl/node/pkg/db"
	"bisonai.com/orakl/node/pkg/raft"
	"bisonai.com/orakl/node/pkg/utils"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/rs/zerolog/log"
)

func NewNode(h host.Host, ps *pubsub.PubSub, topicString string) (*AggregatorNode, error) {
	topic, err := ps.Join(topicString)
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	leaderTimeout := 5 * time.Second

	aggregator := AggregatorNode{
		Raft: raft.NewRaftNode(h, ps, topic, sub, 100), // consider updating after testing

		LeaderJobTimeout: &leaderTimeout,

		CollectedPrices: map[int64][]int64{},
		AggregatorMutex: sync.Mutex{},
	}

	return &aggregator, nil
}

func (n *AggregatorNode) Run(ctx context.Context) {
	n.Raft.Run(ctx, n)
}

func (n *AggregatorNode) GetLeaderJobTimeout() *time.Duration {
	return n.LeaderJobTimeout
}

func (n *AggregatorNode) GetLeaderJobTicker() *time.Ticker {
	return n.LeaderJobTicker
}

func (n *AggregatorNode) SetLeaderJobTicker(d *time.Duration) error {
	if d == nil {
		n.LeaderJobTicker = nil
		return nil
	}
	n.LeaderJobTicker = time.NewTicker(*d)
	return nil
}

func (n *AggregatorNode) LeaderJob() error {
	// leader continously sends roundId in regular basis and triggers all other nodes to run its job
	n.RoundID++
	roundMessage := RoundSyncMessage{
		LeaderID: n.Raft.Host.ID().String(),
		RoundID:  n.RoundID,
	}

	marshalledRoundMessage, err := json.Marshal(roundMessage)
	if err != nil {
		return err
	}

	message := raft.Message{
		Type:     RoundSync,
		SentFrom: n.Raft.Host.ID().String(),
		Data:     json.RawMessage(marshalledRoundMessage),
	}

	return n.Raft.PublishMessage(message)
}

func (n *AggregatorNode) HandleCustomMessage(message raft.Message) error {
	switch message.Type {
	case RoundSync:
		return n.HandleRoundSyncMessage(message)
		// every node runs its job when leader sends roundSync message
	case PriceData:
		return n.HandlePriceDataMessage(message)
	}
	return nil
}

/*
TODO: adding another phase to agree on roundId

1. leader sends roundSync message
2. followers check if the leader's roundId is greater than its own roundId
3. if it is, follower will send signal to leader to update roundId
*/
func (n *AggregatorNode) HandleRoundSyncMessage(msg raft.Message) error {
	var roundSyncMessage RoundSyncMessage
	err := json.Unmarshal(msg.Data, &roundSyncMessage)
	if err != nil {
		return err
	}
	n.RoundID = roundSyncMessage.RoundID

	// pull latest local aggregate and send to peers
	// latestAggregate := utils.RandomNumberGenerator()
	var updateValue int64
	value, updateTime, err := n.getLatestLocalAggregate(n.nodeCtx)
	if err != nil {
		return err
	}

	// when local aggregate have not been updated, send priceDataMessage with -1 value indicating it should be ignored
	if !n.LastLocalAggregateTime.IsZero() && n.LastLocalAggregateTime.Equal(updateTime) {
		updateValue = -1
	} else {
		updateValue = value
		n.LastLocalAggregateTime = updateTime
	}

	priceDataMessage := PriceDataMessage{
		RoundID:   n.RoundID,
		PriceData: updateValue,
	}

	marshalledPriceDataMessage, err := json.Marshal(priceDataMessage)
	if err != nil {
		return err
	}

	message := raft.Message{
		Type:     PriceData,
		SentFrom: n.Raft.Host.ID().String(),
		Data:     json.RawMessage(marshalledPriceDataMessage),
	}

	return n.Raft.PublishMessage(message)
}

func (n *AggregatorNode) HandlePriceDataMessage(msg raft.Message) error {
	var priceDataMessage PriceDataMessage
	err := json.Unmarshal(msg.Data, &priceDataMessage)
	if err != nil {
		return err
	}
	n.AggregatorMutex.Lock()
	defer n.AggregatorMutex.Unlock()
	if _, ok := n.CollectedPrices[priceDataMessage.RoundID]; !ok {
		n.CollectedPrices[priceDataMessage.RoundID] = []int64{}
	}

	n.CollectedPrices[priceDataMessage.RoundID] = append(n.CollectedPrices[priceDataMessage.RoundID], priceDataMessage.PriceData)
	if len(n.CollectedPrices[priceDataMessage.RoundID]) >= len(n.Raft.Ps.ListPeers(n.Raft.Topic.String()))+1 {
		filteredCollectedPrices := FilterNegative(n.CollectedPrices[priceDataMessage.RoundID])

		// handle aggregation here once all the data have been collected
		median := utils.FindMedianInt64(filteredCollectedPrices)
		log.Debug().Int64("roundId", priceDataMessage.RoundID).Int64("global_aggregate", median).Msg("global aggregated")
		err := n.insertGlobalAggregate(n.nodeCtx, n.Name, median, priceDataMessage.RoundID)
		if err != nil {
			log.Error().Err(err).Msg("failed to insert global aggregate")
			return err
		}
		delete(n.CollectedPrices, priceDataMessage.RoundID)
	}
	return nil
}

func (n *AggregatorNode) getLatestLocalAggregate(ctx context.Context) (int64, time.Time, error) {
	redisAggregate, err := GetLatestLocalAggregateFromRdb(ctx, n.Name)
	if err != nil {
		pgsqlAggregate, err := GetLatestLocalAggregateFromPgs(ctx, n.Name)
		if err != nil {
			return 0, time.Time{}, err
		}
		return pgsqlAggregate.Value, pgsqlAggregate.Timestamp, nil
	}
	return redisAggregate.Value, redisAggregate.Timestamp, nil
}

func (n *AggregatorNode) insertGlobalAggregate(ctx context.Context, name string, value int64, round int64) error {
	_, err := db.QueryRow[globalAggregate](ctx, InsertGlobalAggregateQuery, map[string]any{"name": name, "value": value, "round": round})
	if err != nil {
		return err
	}
	return nil
}

func (n *AggregatorNode) executeDeviation() error {
	// signals for deviation job which triggers immediate aggregation and sends submission request to submitter
	return nil
}