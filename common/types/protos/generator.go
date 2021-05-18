//go:generate pwd

//go:generate protoc -I=$GOPATH/src/:. --gogofaster_out=. messages.proto

package protos //nolint