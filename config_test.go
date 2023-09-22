package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "connectbox-exporter.yml")
		require.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(
			"listen_addr: 0.0.0.0:9119\n" +
				"targets:\n" +
				"  - addr: 192.168.178.1\n" +
				"    username: NULL\n" +
				"    password: password",
		)
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		conf, err := ReadConfig(file.Name())
		require.NoError(t, err)

		want := Config{
			ListenAddr: "0.0.0.0:9119",
			Targets: []Target{{
				Addr:     "192.168.178.1",
				Username: "NULL",
				Password: "password",
			}},
		}
		require.Equal(t, want, conf)
	})

	t.Run("empty listen address", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "connectbox-exporter.yml")
		require.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(
			"targets:\n" +
				"  - addr: 192.168.178.1\n" +
				"    username: NULL\n" +
				"    password: password",
		)
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		conf, err := ReadConfig(file.Name())
		require.NoError(t, err)

		want := Config{
			ListenAddr: "0.0.0.0:9119",
			Targets: []Target{{
				Addr:     "192.168.178.1",
				Username: "NULL",
				Password: "password",
			}},
		}
		require.Equal(t, want, conf)
	})

	t.Run("empty target address", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "connectbox-exporter.yml")
		require.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(
			"listen_addr: 0.0.0.0:9119\n" +
				"targets:\n" +
				"  - username: NULL\n" +
				"    password: password",
		)
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		_, err = ReadConfig(file.Name())
		require.ErrorContains(t, err, "found target with empty address")
	})

	t.Run("empty target username", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "connectbox-exporter.yml")
		require.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(
			"listen_addr: 0.0.0.0:9119\n" +
				"targets:\n" +
				"  - addr: 192.168.178.1\n" +
				"    password: password",
		)
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		conf, err := ReadConfig(file.Name())
		require.NoError(t, err)

		want := Config{
			ListenAddr: "0.0.0.0:9119",
			Targets: []Target{{
				Addr:     "192.168.178.1",
				Username: "NULL",
				Password: "password",
			}},
		}
		require.Equal(t, want, conf)
	})

	t.Run("empty target password", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "connectbox-exporter.yml")
		require.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(
			"listen_addr: 0.0.0.0:9119\n" +
				"targets:\n" +
				"  - addr: 192.168.178.1\n" +
				"    username: NULL",
		)
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		_, err = ReadConfig(file.Name())
		require.ErrorContains(t, err, "found target with empty password")
	})

	t.Run("invalid yaml", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "connectbox-exporter.yml")
		require.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString("hello: world: !")
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		_, err = ReadConfig(file.Name())
		require.ErrorContains(t, err, "unmarshal file")
	})

	t.Run("non-existing config file", func(t *testing.T) {
		_, err := ReadConfig("not-exists.yml")
		require.ErrorContains(t, err, "no such file or directory")
	})
}
