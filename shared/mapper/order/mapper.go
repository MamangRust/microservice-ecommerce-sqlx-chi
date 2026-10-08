package orderapimapper

type OrderResponseMapper interface {
	QueryMapper() OrderQueryResponseMapper
	CommandMapper() OrderCommandResponseMapper
}

type orderResponseMapper struct {
	queryMapper   OrderQueryResponseMapper
	commandMapper OrderCommandResponseMapper
}

func NewOrderResponseMapper() OrderResponseMapper {
	return &orderResponseMapper{
		queryMapper:   NewOrderQueryResponseMapper(),
		commandMapper: NewOrderCommandResponseMapper(),
	}
}

func (o *orderResponseMapper) QueryMapper() OrderQueryResponseMapper {
	return o.queryMapper
}

func (o *orderResponseMapper) CommandMapper() OrderCommandResponseMapper {
	return o.commandMapper
}
