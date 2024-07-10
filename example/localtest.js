import ibmmq from 'k6/x/ibmmq';

const client = ibmmq.newClient()

export default function () {
    const sourceQueue = "DEV.QUEUE.1"
    const replyQueue = "DEV.QUEUE.2"
    const sourceMessage = "My Message"
    const replyMessage = "ReplyMsg"
    // The below parameter enable/disable a simulated application that will consume
    // the message in the source queue and put a reply message in the reply queue
    const simulateReply = true

    const msgID = ibmmq.send(client, sourceQueue, replyQueue, sourceMessage, simulateReply)
    ibmmq.receive(client, replyQueue, msgID, replyMessage)
}
