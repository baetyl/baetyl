package security

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/pki/models"
	"github.com/baetyl/baetyl/store"
	"github.com/stretchr/testify/assert"
	bh "github.com/timshannon/bolthold"
)

func genBolthold(t *testing.T) *bh.Store {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	assert.NotNil(t, f)

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	return sto
}

func genStorageMockCert() models.Cert {
	return models.Cert{
		CertId:      "CertId",
		ParentId:    "ParentId",
		Type:        "Type",
		CommonName:  "CommonName",
		Csr:         "Csr",
		Content:     "Content",
		PrivateKey:  "PrivateKey",
		Description: "Description",
		NotBefore:   time.Unix(1000, 1),
		NotAfter:    time.Unix(1000, 1),
	}
}

func TestBhStorage(t *testing.T) {
	sto := genBolthold(t)
	cli := NewStorage(sto)

	cert := genStorageMockCert()

	// good case
	err := cli.CreateCert(cert)
	assert.NoError(t, err)

	c0, err := cli.GetCert(cert.CertId)
	assert.NoError(t, err)
	assert.EqualValues(t, cert, *c0)

	cert.CommonName = "new CommonName"
	err = cli.UpdateCert(cert)
	assert.NoError(t, err)
	c1, err := cli.GetCert(cert.CertId)
	assert.NoError(t, err)
	assert.EqualValues(t, cert, *c1)

	count, err := cli.CountCertByParentId(cert.ParentId)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	err = cli.DeleteCert(cert.CertId)
	assert.NoError(t, err)
	count, err = cli.CountCertByParentId(cert.ParentId)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	err = cli.Close()
	assert.NoError(t, err)
}
