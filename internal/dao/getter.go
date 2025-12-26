package dao

import (
	"log"
	"sync"
	"tiny-bitly/internal/dao/daoimpls/memory"
	"tiny-bitly/internal/dao/daotypes"
)

type DAOType string

const (
	DAOTypeDatabase DAOType = "database"
	DAOTypeMemory   DAOType = "memory"
)

// The main Data-Access Object (DAO) that contains all entity-specific DAOs.
type DAO struct {
	URLRecordDAO daotypes.URLRecordDAO
}

var (
	memoryDAO     *DAO
	memoryDAOOnce sync.Once
)

// Returns a main DAO containing all entity-specific DAOs of the specified type.
func GetDAOOfType(daoType DAOType) *DAO {
	switch daoType {
	case DAOTypeDatabase:
		// TODO: Implement the database DAO.
		log.Fatalf("database DAO not yet implemented")
		return nil
	case DAOTypeMemory:
		// Return a singleton to ensure that values stored in memory will
		// persist across invocations.
		memoryDAOOnce.Do(func() {
			memoryDAO = &DAO{
				URLRecordDAO: memory.NewURLRecordMemoryDAO(),
			}
		})
		return memoryDAO
	default:
		log.Fatalf("unknown DAO type: %s", daoType)
		return nil
	}
}
