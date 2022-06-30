package tests

import "testing"

func TestPathConfig_ProtonPathConfig(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		s.setFolderPrefix("user", "Folders")
		s.setLabelPrefix("user", "Labels")

		c.C(`A001 CREATE Folders/TestFolder`)
		c.Sx(`A001 OK`)

		c.C(`B001 LIST "" *`)
		c.Sx(`\* LIST .* "Folders"`, `\* LIST .* "Labels"`, `\* LIST .* "INBOX"`, `\* LIST .* "Folders/TestFolder"`)
		c.Sx(`B001 OK`)

		c.C(`A002 CREATE Labels/TestLabel`)
		c.Sx(`A002 OK`)

		c.C(`B002 LIST "" *`)
		c.Sx(`\* LIST .* "Folders"`, `\* LIST .* "Labels"`, `\* LIST .* "INBOX"`, `\* LIST .* "Folders/TestFolder"`,
			`"Labels/TestLabel"`)
		c.Sx(`B002 OK`)

		c.C(`A003 CREATE Invalid/TestFolder`)
		c.Sx(`A003 NO invalid prefix`)

		c.C(`A004 CREATE Folders`)
		c.Sx(`A004 NO a mailbox with that name already exists`)

		c.C(`A005 RENAME Folders/TestFolder NewName`)
		c.Sx(`A005 NO invalid prefix`)

		c.C(`A006 RENAME Folders/TestFolder Labels/TestFolder`)
		c.Sx(`A006 NO rename operation is not allowed`)

		c.C(`A007 RENAME Folders/TestFolder Folders/NewName`)
		c.Sx(`A007 OK`)

		c.C(`A008 SELECT Folders/NewName`)
		c.Sxe(`A008 OK`)

		c.C(`A009 CLOSE`)
		c.Sx(`A009 OK`)

		c.C(`A010 DELETE Folders/NewName`)
		c.Sx(`A010 OK`)

		c.C(`A011 DELETE Folders/A`)
		c.Sx(`A011 NO no such mailbox`)
	})
}

func TestPathConfig_DotDelimiter(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		s.setFolderPrefix("user", "Folders")
		s.setLabelPrefix("user", "Labels")

		c.C(`A001 CREATE Folders.TestFolder`)
		c.Sx(`A001 OK`)

		c.C(`B001 LIST "" *`)
		c.Sx(`\* LIST .* "Folders"`, `\* LIST .* "Labels"`, `\* LIST .* "INBOX"`, `\* LIST .* "Folders.TestFolder"`)
		c.Sx(`B001 OK`)

		c.C(`A002 CREATE Labels.TestLabel`)
		c.Sx(`A002 OK`)

		c.C(`B002 LIST "" *`)
		c.Sx(`\* LIST .* "Folders"`, `\* LIST .* "Labels"`, `\* LIST .* "INBOX"`, `\* LIST .* "Folders\.TestFolder"`,
			`"Labels\.TestLabel"`)
		c.Sx(`B002 OK`)

		c.C(`A003 CREATE Invalid.TestFolder`)
		c.Sx(`A003 NO invalid prefix`)

		c.C(`A004 CREATE Folders`)
		c.Sx(`A004 NO a mailbox with that name already exists`)

		c.C(`A005 RENAME Folders.TestFolder NewName`)
		c.Sx(`A005 NO invalid prefix`)

		c.C(`A006 RENAME Folders.TestFolder Labels.TestFolder`)
		c.Sx(`A006 NO rename operation is not allowed`)

		c.C(`A007 RENAME Folders.TestFolder Folders.NewName`)
		c.Sx(`A007 OK`)

		c.C(`A008 SELECT Folders.NewName`)
		c.Sxe(`A008 OK`)

		c.C(`A009 CLOSE`)
		c.Sx(`A009 OK`)

		c.C(`A010 DELETE Folders.NewName`)
		c.Sx(`A010 OK`)

		c.C(`A011 DELETE Folders.A`)
		c.Sx(`A011 NO no such mailbox`)
	})
}
