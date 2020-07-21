package security

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/pki"
	"github.com/baetyl/baetyl-go/v2/pki/models"
	bh "github.com/timshannon/bolthold"
)

type bhStorage struct {
	sto *bh.Store
}

func NewStorage(sto *bh.Store) pki.Storage {
	return &bhStorage{
		sto: sto,
	}
}

func (s *bhStorage) CreateCert(cert models.Cert) error {
	err := s.sto.Insert(cert.CertId, cert)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *bhStorage) DeleteCert(certId string) error {
	tp := models.Cert{}
	err := s.sto.Delete(certId, tp)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *bhStorage) UpdateCert(cert models.Cert) error {
	err := s.sto.Update(cert.CertId, cert)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *bhStorage) GetCert(certId string) (*models.Cert, error) {
	cert := &models.Cert{}
	err := s.sto.Get(certId, cert)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return cert, nil
}

func (s *bhStorage) CountCertByParentId(parentId string) (int, error) {
	res := []models.Cert{}
	err := s.sto.Find(&res, bh.Where("ParentId").Eq(parentId))
	if err != nil {
		return 0, errors.Trace(err)
	}
	return len(res), nil
}

func (s *bhStorage) Close() error {
	if s.sto == nil {
		return nil
	}
	err := s.sto.Close()
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
