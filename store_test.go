package marlow

import "fmt"
import "io"
import "sync"
import "bytes"
import "net/url"
import "testing"
import "go/ast"
import "go/token"
import "go/parser"
import "github.com/franela/goblin"
import "github.com/marlow/marlow/writing"

type storeTestScaffold struct {
	output   *bytes.Buffer
	imports  chan string
	methods  map[string]writing.FuncDecl
	record   url.Values
	received map[string]bool
	closed   bool
	wg       *sync.WaitGroup
}

func (s *storeTestScaffold) g() io.Reader {
	record := marlowRecord{
		importChannel: s.imports,
		config:        s.record,
	}

	return newStoreGenerator(record, s.methods)
}

func (s *storeTestScaffold) parsed() (*ast.File, error) {
	return parser.ParseFile(token.NewFileSet(), "", s.output, parser.AllErrors)
}

func (s *storeTestScaffold) close() {
	s.closed = true
	close(s.imports)
	s.wg.Wait()
}

func Test_StoreGenerator(t *testing.T) {
	g := goblin.Goblin(t)

	var scaffold *storeTestScaffold

	g.Describe("store generator test suite", func() {

		g.BeforeEach(func() {
			scaffold = &storeTestScaffold{
				output:   new(bytes.Buffer),
				imports:  make(chan string),
				record:   make(url.Values),
				methods:  make(map[string]writing.FuncDecl),
				received: make(map[string]bool),
				closed:   false,
				wg:       &sync.WaitGroup{},
			}

			scaffold.wg.Add(1)

			go func() {
				for i := range scaffold.imports {
					scaffold.received[i] = true
				}
				scaffold.wg.Done()
			}()
		})

		g.AfterEach(func() {
			if scaffold.closed == false {
				scaffold.close()
			}
		})

		g.It("returns an error if the store name is not valid", func() {
			_, e := io.Copy(scaffold.output, scaffold.g())
			g.Assert(e == nil).Equal(false)
		})

		g.It("does not inject any imports if name is invalid", func() {
			io.Copy(scaffold.output, scaffold.g())
			scaffold.close()
			g.Assert(len(scaffold.received)).Equal(0)
		})

		g.Describe("with a valid store name", func() {

			g.BeforeEach(func() {
				scaffold.record.Set("storeName", "BookStore")
				fmt.Fprintln(scaffold.output, "package marlowt")
			})

			g.It("injects fmt and sql packages into import stream", func() {
				io.Copy(scaffold.output, scaffold.g())
				scaffold.close()
				g.Assert(scaffold.received["database/sql"]).Equal(true)
				g.Assert(scaffold.received["io"]).Equal(true)
				g.Assert(scaffold.received["os"]).Equal(true)
				g.Assert(len(scaffold.received)).Equal(3)
			})

			g.It("writes valid golang code if store name is present", func() {
				io.Copy(scaffold.output, scaffold.g())
				_, e := scaffold.parsed()
				g.Assert(e).Equal(nil)
			})
		})
	})
}
