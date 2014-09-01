package storage

import (
	"log"
	"net/http"

	"github.com/coreos/go-etcd/etcd"
	"github.com/juju/errgo"
)

// KeyFormat (ETCD):
//  /moinz.de/userd/user/<userid> = JSON()
//  /moinz.de/userd/email/<email> = userid()
//  /moinz.de/userd/login_name/<login_name> = userid()
//
// =>
//  <prefix>/<index>/<key> = <value>
func etcdKey(prefix, index, key string) string {
	return prefix + "/" + index + "/" + key
}

func NewEtcdStorage(peers []string, prefix string, ttl uint64, syncCluster, logCURL bool, logger *log.Logger) *keyValueStorage {
	client := etcd.NewClient(peers)

	if logger != nil {
		etcd.SetLogger(logger)
	}

	if logCURL {
		go func() {
			for {
				curl := client.RecvCURL()

				log.Println(curl)
			}
		}()
		client.OpenCURL()
	}

	if syncCluster {
		client.SyncCluster()
	}
	return newKeyValueStorage(&EtcdStorageDriver{client, prefix, ttl})
}

type EtcdStorageDriver struct {
	client *etcd.Client
	prefix string
	ttl    uint64
}

func (d *EtcdStorageDriver) Path(index, name string) string {
	return etcdKey(d.prefix, index, name)
}

// Index is called initially to create a helper for accessing an index
func (d *EtcdStorageDriver) Index(name string) keyValueIndex {
	return &EtcdIndex{d, name}
}

// Set writes the json with data
func (d *EtcdStorageDriver) Set(userID, json string) error {
	key := d.Path("user", userID)
	_, err := d.client.Set(key, json, d.ttl)
	return errgo.Mask(err)
}

func (d *EtcdStorageDriver) create(key, value string) error {
	_, err := d.client.Create(key, string(value), d.ttl)
	return errgo.Mask(err)
}

// Lookup returns the json previously written with Set().
func (d *EtcdStorageDriver) Lookup(userID string) (string, bool, error) {
	json, ok, err := d.lookupIndex("user", userID)
	if err != nil {
		return "", false, errgo.Mask(err)
	}
	return json, ok, nil
}

func (d *EtcdStorageDriver) lookupIndex(index, key string) (string, bool, error) {
	path := d.Path(index, key)

	rawResp, err := d.client.RawGet(path, false, false)
	if err != nil {
		return "", false, errgo.Mask(err)
	}
	if rawResp.StatusCode == http.StatusNotFound {
		return "", false, nil
	}

	resp, err := rawResp.Unmarshal()
	if err != nil {
		return "", false, errgo.Mask(err)
	}

	return resp.Node.Value, true, nil
}

func (d *EtcdStorageDriver) removeIndex(index, key string) error {
	path := d.Path(index, key)
	_, err := d.client.Delete(path, false)
	if err != nil {
		return errgo.Mask(err)
	}
	return nil
}

// EtcdIndex implements KeyValueIndex on Etcd.
type EtcdIndex struct {
	Storage *EtcdStorageDriver
	Name    string
}

func (s *EtcdIndex) Put(key, userID string) error {
	path := s.Storage.Path(s.Name, key)
	return errgo.Mask(s.Storage.create(path, userID))
}

func (s *EtcdIndex) Remove(key string) error {
	if err := s.Storage.removeIndex(s.Name, key); err != nil {
		return errgo.Mask(err)
	}
	return nil
}

func (s *EtcdIndex) Lookup(key string) (string, bool, error) {
	json, ok, err := s.Storage.lookupIndex(s.Name, key)
	if err != nil {
		return "", false, errgo.Mask(err)
	}
	return json, ok, nil
}
