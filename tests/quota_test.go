package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
)

func TestGetQuota(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		s.setQuota("user", "", imap.QuotaResource{
			ResourceName: "STORAGE",
			Usage:        100,
			Limit:        512,
		})

		c.C(`A001 GETQUOTA ""`)
		c.S(`* QUOTA "" (STORAGE 100 512)`)
		c.S(`A001 OK GETQUOTA`)
	})
}

func TestGetQuotaMultipleResources(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		s.setQuota("user", "user-root", imap.QuotaResource{
			ResourceName: "STORAGE",
			Usage:        200,
			Limit:        1024,
		}, imap.QuotaResource{
			ResourceName: "MESSAGE",
			Usage:        50,
			Limit:        1000,
		})

		c.C(`A001 GETQUOTA "user-root"`)
		c.S(`* QUOTA "user-root" (STORAGE 200 1024 MESSAGE 50 1000)`)
		c.S(`A001 OK GETQUOTA`)
	})
}

func TestGetQuotaRoot(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		s.setQuota("user", "", imap.QuotaResource{
			ResourceName: "STORAGE",
			Usage:        100,
			Limit:        512,
		})

		c.C(`A001 GETQUOTAROOT INBOX`)
		c.S(`* QUOTAROOT INBOX ""`)
		c.S(`* QUOTA "" (STORAGE 100 512)`)
		c.S(`A001 OK GETQUOTAROOT`)
	})
}

func TestGetQuotaRootWithNamedRoot(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		s.setQuota("user", "my-root", imap.QuotaResource{
			ResourceName: "STORAGE",
			Usage:        300,
			Limit:        2048,
		})
		s.setMailboxQuotaRoot("user", "INBOX", "my-root")

		c.C(`A001 GETQUOTAROOT INBOX`)
		c.S(`* QUOTAROOT INBOX "my-root"`)
		c.S(`* QUOTA "my-root" (STORAGE 300 2048)`)
		c.S(`A001 OK GETQUOTAROOT`)
	})
}

func TestGetQuotaRootMultipleRoots(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		s.setQuota("user", "root-a", imap.QuotaResource{
			ResourceName: "STORAGE",
			Usage:        10,
			Limit:        100,
		})
		s.setQuota("user", "root-b", imap.QuotaResource{
			ResourceName: "MESSAGE",
			Usage:        5,
			Limit:        50,
		})
		s.setMailboxQuotaRoot("user", "INBOX", "root-a", "root-b")

		c.C(`A001 GETQUOTAROOT INBOX`)
		c.S(`* QUOTAROOT INBOX "root-a" "root-b"`)
		c.S(`* QUOTA "root-a" (STORAGE 10 100)`)
		c.S(`* QUOTA "root-b" (MESSAGE 5 50)`)
		c.S(`A001 OK GETQUOTAROOT`)
	})
}
