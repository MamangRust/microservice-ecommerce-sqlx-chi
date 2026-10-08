package transactionapimapper

type TransactionResponseMapper interface {
	QueryMapper() TransactionQueryResponseMapper
	CommandMapper() TransactionCommandResponseMapper
}

type transactionResponseMapper struct {
	queryMapper   TransactionQueryResponseMapper
	commandMapper TransactionCommandResponseMapper
}

func NewTransactionResponseMapper() TransactionResponseMapper {
	return &transactionResponseMapper{
		queryMapper:   NewTransactionQueryResponseMapper(),
		commandMapper: NewTransactionCommandResponseMapper(),
	}
}

func (t *transactionResponseMapper) QueryMapper() TransactionQueryResponseMapper {
	return t.queryMapper
}

func (t *transactionResponseMapper) CommandMapper() TransactionCommandResponseMapper {
	return t.commandMapper
}
