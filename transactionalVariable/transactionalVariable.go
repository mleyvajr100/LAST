package transactionalvariable

import (
	"fmt"

	"github.com/last/client"
)

type TransactionalVariable interface {
	Get() int32
	Set(int32)
	Name() string
}

type TxVar struct {
	variable string
	value    int32
}

func (txVar *TxVar) Get() int32 {
	fmt.Println("IM herer")
	return client.GetVariable(txVar.variable)
}

func (txVar *TxVar) Set(value int32) {
	client.SetVariable(txVar.variable, value)
}

func (txVar *TxVar) Name() string {
	return txVar.variable
}

func New(variable string, value int32) *TxVar {
	client.SetVariable(variable, value)
	return &TxVar{variable: variable, value: value}
}
