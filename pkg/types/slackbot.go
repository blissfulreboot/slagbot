package types

type OutgoingSlackMessage struct {
	Channel   *string
	UserEmail *string
	Message   string
}
