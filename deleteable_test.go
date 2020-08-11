package marlow

import "io"
import "sync"
import "bytes"
import "net/url"
import "testing"
import "github.com/franela/goblin"
import "github.com/marlow/marlow/writing"
import "github.com/marlow/marlow/constants"

type deleteableTestScaffold struct {
	buffer *bytes.Buffer

	imports chan string
	methods chan writing.FuncDecl
	record  url.Values
	fields  map[string]url.Values

	received map[string]bool
	closed   bool
	wg       *sync.WaitGroup
}

func (s *deleteableTestScaffold) g() io.Reader {
	record := marlowRecord{
		config:        s.record,
		fields:        s.fields,
		importChannel: s.imports,
		storeChannel:  s.methods,
	}
	return newDeleteableGenerator(record)
}

func Test_Deleteable(t *testing.T) {
	g := goblin.Goblin(t)

	var scaffold *deleteableTestScaffold

	g.Describe("deleteable feature generator test suite", func() {

		g.BeforeEach(func() {
			scaffold = &deleteableTestScaffold{
				buffer:   new(bytes.Buffer),
				imports:  make(chan string),
				methods:  make(chan writing.FuncDecl),
				record:   make(url.Values),
				fields:   make(map[string]url.Values),
				received: make(map[string]bool),
				closed:   false,
				wg:       &sync.WaitGroup{},
			}

			scaffold.wg.Add(2)

			go func() {
				for i := range scaffold.imports {
					scaffold.received[i] = true
				}
				scaffold.wg.Done()
			}()

			go func() {
				for range scaffold.methods {
				}
				scaffold.wg.Done()
			}()
		})

		g.AfterEach(func() {
			if scaffold.closed == false {
				close(scaffold.imports)
				close(scaffold.methods)
				scaffold.wg.Wait()
			}
		})

		g.Describe("with a valid record config", func() {

			g.BeforeEach(func() {
				scaffold.record.Set(constants.RecordNameConfigOption, "Author")
				scaffold.record.Set(constants.TableNameConfigOption, "authors")
				scaffold.record.Set(constants.UpdateFieldMethodPrefixConfigOption, "Update")
				scaffold.record.Set(constants.StoreNameConfigOption, "AuthorStore")

				scaffold.fields["ID"] = url.Values{
					"type": []string{"int"},
				}

				scaffold.fields["Name"] = url.Values{
					"type": []string{"string"},
				}

				scaffold.fields["UniversityID"] = url.Values{
					"type": []string{"sql.NullInt64"},
				}
			})

			g.It("generates valid golang", func() {
				_, e := io.Copy(scaffold.buffer, scaffold.g())
				g.Assert(e).Equal(nil)
			})

		})

	})
}
