import { job } from '../src/worker/vrf'
import { IVrfConfig, IVrfListenerWorker } from '../src/types'
import { buildMockLogger } from '../src/logger'
import { VRF_FULFILL_GAS_MINIMUM } from '../src/settings'
import { QUEUE } from './utils'

const vrfConfig: IVrfConfig = {
  sk: 'b368e407363d7435903b1511025bc8345b76aa4fcfe7ab36fb8a71349e1fe95a',
  pk: '045012f4b244b7875e34ac2af0856e463c2c5c94fe754e90a844798047aa32ae34a546a783546ca30b6044ce9612c2d458815249a5f460eb3ce878adf1dc55dec7',
  pkX: '36218519966043180833110848345962110858668389778776319719451647092795737550388',
  pkY: '74756455443057291531071945062373943175438004978429662040378909820867954728647',
  keyHash: '0x1833807c931ca83e42ada8a2730626cdd00871e3013927a2b89f94e82a6844dd'
}

describe('VRF Worker', function () {
  it('Composability test', async function () {
    const logger = buildMockLogger()
    const wrapperFn = await job(QUEUE, vrfConfig, logger)

    const callbackAddress = '0xccf9a654c878848991e46ab23d2ad055ca827979' // random address
    const sender = '0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266' // Hardhat Account #0

    const listenerData: IVrfListenerWorker = {
      callbackAddress,
      blockNum: '1',
      blockHash: '0x0000000000000000000000000000000000000000000000000000000000000000',
      requestId: '1',
      seed: '0',
      accId: '0',
      callbackGasLimit: 2500000,
      numWords: 1,
      sender,
      isDirectPayment: false
    }

    const tx = await wrapperFn({
      data: listenerData
    })

    expect(tx?.payload).toBe(
      '0x4682fdc35012f4b244b7875e34ac2af0856e463c2c5c94fe754e90a844798047aa32ae34a546a783546ca30b6044ce9612c2d458815249a5f460eb3ce878adf1dc55dec77dfe054a2cb1207abfbf6838ddd03967b07cafa2980407c5e0216e9552ab453a0880eb86ccbcef12f564d56e0ace41c6979f5e7568f2d5b130aabf937c2037070000000000000000000000000000000099cc3fd959940a23ca5e13e3e370a31a7c5f4a1bfdf496d4e04d5c27fe574ab3676ef058cc10b4e66f7522b8edf52e62000000000000000000000000000000000000000000000000000000000000000065d021f3017d3e6930cea4ecf43f16ce81460276888e794079dfcf1dd806a0e8d409d5675c4996b0f085b41140f6ffeb55f5ffd356f217a7dc652734d1d3241886f4cb8ca876d3884615fc35e8316dd82ed094596475ff3a86a7d54c76661196651820c7d28d0a321e485f5083bcba8c4b410940e843a91ac61859706e88c3bb37cc48eae15b5682bd4f3899e93f9ec1e11d6eb4a7db56f79d20a7a88fcfb87f51bda72d639250f730c3435a13272f092aefe956c786963ecd643903d8e3db5d0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002625a00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb922660000000000000000000000000000000000000000000000000000000000000000'
    )
    expect(tx?.gasLimit).toBe(
      VRF_FULFILL_GAS_MINIMUM + 10_000 * listenerData.numWords + listenerData.callbackGasLimit
    )
    expect(tx?.to).toBe(callbackAddress)
  })
})
