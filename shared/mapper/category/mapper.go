package categoryapimapper

type CategoryResponseMapper interface {
	QueryMapper() CategoryQueryResponseMapper
	CommandMapper() CategoryCommandResponseMapper
}

type categoryResponseMapper struct {
	queryMapper   CategoryQueryResponseMapper
	commandMapper CategoryCommandResponseMapper
}

func NewCategoryResponseMapper() CategoryResponseMapper {
	return &categoryResponseMapper{
		queryMapper:   NewCategoryQueryResponseMapper(),
		commandMapper: NewCategoryCommandResponseMapper(),
	}
}

func (m *categoryResponseMapper) QueryMapper() CategoryQueryResponseMapper {
	return m.queryMapper
}

func (m *categoryResponseMapper) CommandMapper() CategoryCommandResponseMapper {
	return m.commandMapper
}
