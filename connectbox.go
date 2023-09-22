package main

import "context"

// ConnectBox is a ConnectBox router client, that gets metrics from
// a remote source.
type ConnectBox interface {
	Login(ctx context.Context) error
	Logout(ctx context.Context) error
	Get(ctx context.Context, fn string, out any) error
}
