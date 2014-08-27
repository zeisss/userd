package storage

import (
	"../user"

	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/go-etcd/etcd"
	"github.com/juju/errgo"
)

func NewEtcdStorage(peer, prefix string, ttl uint64) *etcdStorage {
	client := etcd.NewClient([]string{peer})

	client.SyncCluster()

	return &etcdStorage{client, prefix, ttl}
}

type etcdStorage struct {
	client *etcd.Client
	prefix string
	ttl    uint64
}

func (s *etcdStorage) Get(userID string) (user.User, error) {
	var result user.User

	rawResp, err := s.client.RawGet(s.Path(userID), false, false)
	if err != nil {
		return result, errgo.Mask(err)
	}

	if rawResp.StatusCode == http.StatusNotFound {
		return result, UserNotFound
	}

	resp, err := rawResp.Unmarshal()
	if err != nil {
		return result, errgo.Mask(err)
	}
	if resp.Action != "get" {
		return result, fmt.Errorf("Unexpected response from etcd: action=%s", resp.Action)
	}

	s.Unmarshal(resp.Node.Value, &result)

	return result, nil
}

func (s *etcdStorage) Save(user user.User) error {
	_, err := s.client.Set(s.Path(user.ID), s.Marshal(&user), s.ttl)
	return errgo.Mask(err)
}

func (s *etcdStorage) Path(userID string) string {
	return fmt.Sprintf("/%s/user/%s", s.prefix, userID)
}

func (s *etcdStorage) Marshal(user *user.User) string {
	data, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (s *etcdStorage) Unmarshal(value string, user *user.User) {
	if err := json.Unmarshal([]byte(value), user); err != nil {
		panic(err)
	}
}

func (s *etcdStorage) FindByLoginName(loginName string) (user.User, error) {
	result := user.User{}

	return result, nil
}
