package migration

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/fabric8-services/fabric8-common/resource"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	config "github.com/fabric8-services/fabric8-common/configuration"
)

// fail - as t.Fatalf() is not goroutine safe, this function behaves like t.Fatalf().
func fail(t *testing.T, template string, args ...interface{}) {
	fmt.Printf(template, args...)
	fmt.Println()
	t.Fail()
}

func TestConcurrentMigrations(t *testing.T) {
	resource.Require(t, resource.Database)

	configuration, err := config.New("../config.yaml")
	if err != nil {
		panic(fmt.Errorf("Failed to setup the configuration: %s", err.Error()))
	}

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			db, err := sql.Open("postgres", configuration.GetPostgresConfigString())
			if err != nil {
				fail(t, "Cannot connect to DB: %s\n", err)
			}
			err = Migrate(db, configuration.GetPostgresDatabase())
			assert.Nil(t, err)
		}()

	}
	wg.Wait()
}
