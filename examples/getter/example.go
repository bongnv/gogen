package getter

type Example struct {
	Number    int
	String    string
	StringPtr *string `gogen:"skip"`
}
