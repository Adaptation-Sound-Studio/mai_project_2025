package fact

type FactRepository interface {
	Insert(fact *ListenFact) error
}
