package disk_test

import (
	"context"
	"testing"

	"github.com/ilyabrin/disk"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Server", func() {
	var server *ghttp.Server
	var ctx context.Context

	// msg := "Hi there, the end point is :"

	apiClient := disk.New("access_token")

	BeforeEach(func() {
		// start a test http server
		server = ghttp.NewServer()
		ctx = context.Background()
	})
	AfterEach(func() {
		server.Close()
	})

	Context("When GET /disk via GetDiskInfo", func() {
		BeforeEach(func() {
			server.AppendHandlers()
		})
		It("Returns the JSON response with 200 Ok", func() {
			// resp, err := http.Get(server.URL() + "/disk/info")
			diskInfo, err := apiClient.DiskInfo(ctx)
			Expect(err).ShouldNot(HaveOccurred())
			// Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			// body, err := ioutil.ReadAll(resp.Body)
			// resp.Body.Close()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(diskInfo.IsPaid).To(Equal(true))
		})
	})

	// Context("When get request is sent to hello path", func() {
	// 	BeforeEach(func() {
	// 		server.AppendHandlers(
	// 			Handler,
	// 		)
	// 	})
	// 	It("Returns the empty path", func() {
	// 		resp, err := http.Get(server.URL() + "/hello")
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	// 		body, err := ioutil.ReadAll(resp.Body)
	// 		resp.Body.Close()
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		Expect(string(body)).To(Equal(msg + "hello!"))
	// 	})
	// })

	// Context("When get request is sent to read path but there is no file", func() {
	// 	BeforeEach(func() {
	// 		server.AppendHandlers(
	// 			ReadHandler,
	// 		)
	// 	})
	// 	It("Returns internal server error", func() {
	// 		resp, err := http.Get(server.URL() + "/read")
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		Expect(resp.StatusCode).Should(Equal(http.StatusInternalServerError))
	// 		body, err := ioutil.ReadAll(resp.Body)
	// 		resp.Body.Close()
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		Expect(string(body)).To(Equal("open data.txt: no such file or directory\n"))
	// 	})
	// })

	// Context("When get request is sent to read path but file exists", func() {
	// 	BeforeEach(func() {
	// 		file, err := os.Create("data.txt")
	// 		Expect(err).NotTo(HaveOccurred())
	// 		file.Write([]byte("Hi there!"))
	// 		server.AppendHandlers(
	// 			ReadHandler,
	// 		)
	// 	})
	// 	AfterEach(func() {
	// 		err := os.Remove("data.txt")
	// 		Expect(err).NotTo(HaveOccurred())
	// 	})
	// 	It("Reads data from file successfully", func() {
	// 		resp, err := http.Get(server.URL() + "/read")
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		Expect(resp.StatusCode).Should(Equal(http.StatusOK))
	// 		body, err := ioutil.ReadAll(resp.Body)
	// 		resp.Body.Close()
	// 		Expect(err).ShouldNot(HaveOccurred())
	// 		Expect(string(body)).To(Equal("Content in file is...\r\nHi there!"))
	// 	})
	// })
})

func TestDisk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Disk Suite")
}
