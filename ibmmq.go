package xk6ibmmq

import (
	"strconv"

	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/walles/env"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/ibmmq", new(Ibmmq))
}

type Ibmmq struct{}

func (*Ibmmq) NewClient() mqjms.ConnectionFactoryImpl {
	QMName := env.MustGet("MQ_QMGR", env.String)
	Hostname := env.MustGet("MQ_HOST", env.String)
	PortNumber := env.MustGet("MQ_PORT", strconv.Atoi)
	ChannelName := env.MustGet("MQ_CHANNEL", env.String)
	UserName := env.MustGet("MQ_USERID", env.String)
	Password := env.MustGet("MQ_PASSWORD", env.String)

	cf := mqjms.ConnectionFactoryImpl{
		QMName:      QMName,
		Hostname:    Hostname,
		PortNumber:  PortNumber,
		ChannelName: ChannelName,
		UserName:    UserName,
		Password:    Password,
	}

	return cf
}

func (s *Ibmmq) Send(ibmmqClient mqjms.ConnectionFactoryImpl, sendQueueName string, replyQueueName string, msgText string, sim bool) string {
	ctx, errCtx := ibmmqClient.CreateContext()

	if errCtx != nil {
		panic("Error during CreateContext: " + errCtx.GetReason())
	}

	defer ctx.Close()

	sendQueue := ctx.CreateQueue(sendQueueName)
	replyQueue := ctx.CreateQueue(replyQueueName)

	msg := ctx.CreateTextMessageWithString(msgText)

	errReply := msg.SetJMSReplyTo(replyQueue)
	if errReply != nil {
		panic("Error setting reply: " + errReply.GetReason())
	}

	errSend := ctx.CreateProducer().Send(sendQueue, msg)

	if errSend != nil {
		panic("Error sending: " + errSend.GetReason())
	}

	msgID := msg.GetJMSMessageID()

	if sim {
		replyToMessage(ibmmqClient, sendQueueName)
	}

	return msgID
}

func (s *Ibmmq) Receive(ibmmqClient mqjms.ConnectionFactoryImpl, replyQueueName string, msgID string, msgText string) {
	ctx, errCtx := ibmmqClient.CreateContext()

	if errCtx != nil {
		panic("Error during CreateContext: " + errCtx.GetReason())
	}

	defer ctx.Close()

	replyQueue := ctx.CreateQueue(replyQueueName)

	// Receive the reply message, selecting by CorrelID
	requestConsumer, errRConn := ctx.CreateConsumerWithSelector(replyQueue, "JMSCorrelationID = '"+msgID+"'")

	if requestConsumer != nil {
		defer requestConsumer.Close()
	} else {
		panic("Unable to select message: " + errRConn.GetReason())
	}

	rcvMsg, errRvc := requestConsumer.ReceiveNoWait()
	if rcvMsg == nil && errRvc == nil {
		panic("Queue is empty!")
	}

	switch msg := rcvMsg.(type) {
	case jms20subset.TextMessage:
		if msgText != *msg.GetText() {
			panic("Not the reply we expect!")
		}
	default:
		panic("Got something other than a text message: " + errRvc.GetReason())
	}
}

/*
 * Simulate another application replying to a message.
 */
func replyToMessage(cf mqjms.ConnectionFactoryImpl, sendQueueName string) {
	ctx, errCtx := cf.CreateContext()

	if errCtx != nil {
		panic("Error during CreateContext: " + errCtx.GetReason())
	}

	defer ctx.Close()

	sendQueue := ctx.CreateQueue(sendQueueName)

	// Receive the sent message, selecting by CorrelID
	requestConsumer, errRConn := ctx.CreateConsumer(sendQueue)

	if requestConsumer != nil {
		defer requestConsumer.Close()
	} else {
		panic("Error during reading msg: " + errRConn.GetReason())
	}

	rcvMsg, errRvc := requestConsumer.ReceiveNoWait()

	if rcvMsg == nil && errRvc == nil {
		panic("Queue is empty!")
	}

	reqMsgID := rcvMsg.GetJMSMessageID()
	replyDest := rcvMsg.GetJMSReplyTo()

	switch rcvMsg.(type) {
	case jms20subset.TextMessage:
	default:
		panic("Got something other than a text message: " + errRvc.GetReason())
	}

	// Reply to the sent message, set CorrelID
	replyMsgBody := "ReplyMsg"
	replyMsg := ctx.CreateTextMessageWithString(replyMsgBody)
	errJMSCorr := replyMsg.SetJMSCorrelationID(reqMsgID)

	if errJMSCorr != nil {
		panic("Error setting correlation: " + errJMSCorr.GetReason())
	}

	// Send the reply message back to the reply queue
	errSend := ctx.CreateProducer().Send(replyDest, replyMsg)

	if errSend != nil {
		panic("Error sending: " + errSend.GetReason())
	}
}
