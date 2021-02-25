//go:generate pwd

//go:generate protoc --gogofaster_out=. message.proto

//go:generate protoc -I=$GOPATH/src/:. --gogofaster_out=. transaction.proto

package protos //nolint