package config

const (
	InsertConfigQuery     = "INSERT INTO configs (name, address, fetch_interval, aggregate_interval, submit_interval) VALUES (@name, @address, @fetch_interval, @aggregate_interval, @submit_interval) RETURNING *"
	SelectConfigQuery     = "SELECT * FROM configs"
	SelectConfigByIdQuery = "SELECT * FROM configs WHERE id = @id"
	DeleteConfigQuery     = "DELETE FROM configs WHERE id = @id RETURNING *"
	InsertFeedQuery       = "INSERT INTO feeds (name, definition, config_id) VALUES (@name, @definition, @config_id)"
)