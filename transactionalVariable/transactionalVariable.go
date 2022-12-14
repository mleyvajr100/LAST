package transactionalvariable

import (
	"github.com/last/client"
)

type TransactionalVariable interface {
	Get(string) int32
	Set(int32, string)
	Name() string
}

type TxVar struct {
	variable string
	value    int32
}

func CreateSession() string {
	return client.CreateSession()
}

func CommitSession(sessionID string) {
	client.CommitSession(sessionID)
}

func (txVar *TxVar) Get(sessionID string) int32 {
	return client.GetVariable(txVar.variable, sessionID)
}

func (txVar *TxVar) Set(value int32, sessionID string) {
	client.SetVariable(txVar.variable, value, sessionID)
}

func (txVar *TxVar) Name() string {
	return txVar.variable
}

func New(variable string, value int32, sessionID string) *TxVar {
	client.SetVariable(variable, value, sessionID)
	return &TxVar{variable: variable, value: value}
}
