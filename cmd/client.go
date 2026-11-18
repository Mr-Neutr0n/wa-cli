package cmd

import (
	"context"

	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/codebyoketch/wa-cli/internal/chatstore"
	waerrors "github.com/codebyoketch/wa-cli/internal/errors"
	"github.com/codebyoketch/wa-cli/internal/msgstore"
	"github.com/codebyoketch/wa-cli/internal/store"
	"github.com/codebyoketch/wa-cli/internal/whatsapp"
)

// libLog is the logger handed to whatsmeow: warnings and errors only.
func libLog() waLog.Logger { return waLog.Stdout("Database", "WARN", true) }

// openClient opens the session store and builds a client on it, logging
// to log. cs and ms may be nil; when set, the client records the chats
// and messages it sees in them.
func openClient(ctx context.Context, log waLog.Logger, cs *chatstore.Store, ms *msgstore.Store) (*whatsapp.Client, error) {
	container, err := store.Open(ctx, a.Config.DataDir, libLog())
	if err != nil {
		return nil, err
	}
	return whatsapp.New(ctx, container, log, cs, ms)
}

// loggedInClient is openClient for commands that need a linked account.
func loggedInClient(ctx context.Context, cs *chatstore.Store, ms *msgstore.Store) (*whatsapp.Client, error) {
	client, err := openClient(ctx, libLog(), cs, ms)
	if err != nil {
		return nil, err
	}
	if !client.IsLoggedIn() {
		return nil, waerrors.ErrNotLoggedIn
	}
	return client, nil
}

// connectedClient is loggedInClient, connected. Callers must defer
// client.Disconnect().
func connectedClient(ctx context.Context, cs *chatstore.Store, ms *msgstore.Store) (*whatsapp.Client, error) {
	client, err := loggedInClient(ctx, cs, ms)
	if err != nil {
		return nil, err
	}
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

// connectForSend is connectedClient with the local stores attached, so a
// sent message is recorded. It also returns the message store.
func connectForSend(ctx context.Context) (*whatsapp.Client, *msgstore.Store, error) {
	ms := msgstore.New(a.Config.DataDir)
	client, err := connectedClient(ctx, chatstore.New(a.Config.DataDir), ms)
	return client, ms, err
}
