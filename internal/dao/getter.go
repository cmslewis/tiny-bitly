package dao

import (
	"log"
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

// Returns a main DAO containing all entity-specific DAOs of the specified type.
func GetDAOOfType(daoType DAOType) *DAO {
	switch daoType {
	case DAOTypeDatabase:
		// TODO: Implement the database DAO.
		log.Fatalf("database DAO not yet implemented")
		return nil
	case DAOTypeMemory:
		return &DAO{
			URLRecordDAO: memory.NewURLRecordMemoryDAO(),
		}
	default:
		log.Fatalf("unknown DAO type: %s", daoType)
		return nil
	}
}
