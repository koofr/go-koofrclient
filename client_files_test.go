package koofrclient_test

import (
	"bytes"
	"io/ioutil"

	k "github.com/koofr/go-koofrclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClientFiles", func() {
	It("should get file info", func() {
		info, err := client.FilesInfo(defaultMountId, "/")
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Name).To(Equal(""))
	})

	It("should list files", func() {
		files, err := client.FilesList(defaultMountId, "/")
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(0))
	})

	It("should get files tree", func() {
		tree, err := client.FilesTree(defaultMountId, "/")
		Expect(err).NotTo(HaveOccurred())
		Expect(tree.Name).To(Equal(""))
		Expect(tree.Children).To(HaveLen(0))
	})

	It("should create new folder and delete it", func() {
		err := client.FilesNewFolder(defaultMountId, "/", "dir")
		Expect(err).NotTo(HaveOccurred())
		err = client.FilesDelete(defaultMountId, "/dir")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.FilesInfo(defaultMountId, "/dir")
		Expect(err).To(HaveOccurred())
	})

	It("should copy a folder", func() {
		_ = client.FilesDelete(defaultMountId, "/dir")     // cleanup
		_ = client.FilesDelete(defaultMountId, "/dircopy") // cleanup

		err := client.FilesNewFolder(defaultMountId, "/", "dir")
		Expect(err).NotTo(HaveOccurred())
		err = client.FilesCopy(defaultMountId, "/dir", defaultMountId, "/dircopy")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.FilesInfo(defaultMountId, "/dir")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.FilesInfo(defaultMountId, "/dircopy")
		Expect(err).NotTo(HaveOccurred())

		_ = client.FilesDelete(defaultMountId, "/dir")     // cleanup
		_ = client.FilesDelete(defaultMountId, "/dircopy") // cleanup
	})

	It("should move a folder", func() {
		_ = client.FilesDelete(defaultMountId, "/dir")      // cleanup
		_ = client.FilesDelete(defaultMountId, "/dirmoved") // cleanup

		err := client.FilesNewFolder(defaultMountId, "/", "dir")
		Expect(err).NotTo(HaveOccurred())
		err = client.FilesMove(defaultMountId, "/dir", defaultMountId, "/dirmoved")
		Expect(err).NotTo(HaveOccurred())
		_, err = client.FilesInfo(defaultMountId, "/dir")
		Expect(err).To(HaveOccurred())
		_, err = client.FilesInfo(defaultMountId, "/dirmoved")
		Expect(err).NotTo(HaveOccurred())

		_ = client.FilesDelete(defaultMountId, "/dirmoved") // cleanup
	})

	It("should put file, get it and delete it", func() {
		_ = client.FilesDelete(defaultMountId, "/file.txt") // cleanup

		newName, err := client.FilesPut(defaultMountId, "/", "file.txt", bytes.NewReader([]byte("content")))
		Expect(err).NotTo(HaveOccurred())
		Expect(newName).To(Equal("file.txt"))
		reader, err := client.FilesGet(defaultMountId, "/file.txt")
		Expect(err).NotTo(HaveOccurred())
		content, err := ioutil.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).To(Equal([]byte("content")))
		reader, err = client.FilesGetRange(defaultMountId, "/file.txt", &k.FileSpan{Start: 2, End: 3})
		Expect(err).NotTo(HaveOccurred())
		content, err = ioutil.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).To(Equal([]byte("nt")))
		err = client.FilesDelete(defaultMountId, "/file.txt")
		Expect(err).NotTo(HaveOccurred())
	})
})
