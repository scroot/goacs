package impl

import (
	".."
	"../../acs/xml"
	"../../models/cpe"
	"../interfaces"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type MysqlCPERepositoryImpl struct {
	db *sql.DB
}

func NewMysqlCPERepository(connection *sql.DB) interfaces.CPERepository {
	return &MysqlCPERepositoryImpl{
		db: connection,
	}
}

func (r *MysqlCPERepositoryImpl) Find(uuid string) (*cpe.CPE, error) {

	return &cpe.CPE{}, nil
}

func (r *MysqlCPERepositoryImpl) FindBySerial(serial string) (*cpe.CPE, error) {
	r.db.Ping()

	result, err := r.db.Query("SELECT uuid, serial_number, hardware_version FROM cpe WHERE serial_number=? LIMIT 1", serial)

	if err != nil {
		fmt.Println("Error while fetching query results")
		fmt.Println(err.Error())
	}

	for result.Next() {
		cpeInstance := new(cpe.CPE)
		err = result.Scan(&cpeInstance.UUID, &cpeInstance.SerialNumber, &cpeInstance.HardwareVersion)
		if err != nil {
			fmt.Println("Error while fetching query results")
			fmt.Println(err.Error())
		}
		return cpeInstance, nil
	}

	return nil, repository.ErrNotFound
}

func (r *MysqlCPERepositoryImpl) Create(cpe *cpe.CPE) (bool, error) {
	r.db.Ping()

	uuidInstance, _ := uuid.NewRandom()
	uuidString := uuidInstance.String()

	_, err := r.db.Exec(`INSERT INTO cpe SET uuid=?, 
			serial_number=?, 
			hardware_version=?, 
			software_version=?, 
            connection_request_url=?,
			created_at=?, 
			updated_at=?
			`,
		uuidString,
		cpe.SerialNumber,
		cpe.HardwareVersion,
		cpe.SoftwareVersion,
		cpe.ConnectionRequestUrl,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		fmt.Println(err)
		return false, repository.ErrInserting
	}

	cpe.UUID = uuidInstance.String()

	return true, nil
}

func (r *MysqlCPERepositoryImpl) UpdateOrCreate(cpe *cpe.CPE) (result bool, err error) {

	dbCPE, _ := r.FindBySerial(cpe.SerialNumber)

	if dbCPE == nil {
		result, err = r.Create(cpe)
	} else {
		fmt.Println("Updating CPE")
		stmt, _ := r.db.Prepare(`UPDATE cpe SET 
               hardware_version=?, 
               software_version=?, 
               connection_request_url=?, 
               updated_at=? 
			   WHERE uuid=?`)

		_, err := stmt.Exec(
			cpe.HardwareVersion,
			cpe.SoftwareVersion,
			cpe.ConnectionRequestUrl,
			time.Now(),
			dbCPE.UUID,
		)
		cpe.UUID = dbCPE.UUID

		if err != nil {
			return false, repository.ErrUpdating
		}

		result = true
		err = nil

	}

	return result, err
}

func (r *MysqlCPERepositoryImpl) FindParameter(cpe *cpe.CPE, parameterKey string) (*xml.ParameterValueStruct, error) {
	result, err := r.db.Query("SELECT name, value, type  FROM cpe_parameters WHERE cpe_uuid=? AND name=? LIMIT 1", cpe.UUID, parameterKey)

	if err != nil {
		fmt.Println("Error while fetching query results")
		fmt.Println(err.Error())
	}

	for result.Next() {
		parameterValueStruct := new(xml.ParameterValueStruct)
		err = result.Scan(&parameterValueStruct.Name, &parameterValueStruct.Value.Value, &parameterValueStruct.Value.Type)

		if err != nil {
			fmt.Println("Error while fetching query results")
			fmt.Println(err.Error())
		}
		return parameterValueStruct, nil
	}

	return nil, repository.ErrNotFound
}

func (r *MysqlCPERepositoryImpl) CreateParameter(cpe *cpe.CPE, parameter xml.ParameterValueStruct) (bool, error) {
	var query string = `INSERT INTO cpe_parameters (cpe_uuid, name, value, type, flags, created_at, updated_at) 
						VALUES (?, ?, ?, ?, ?, ?, ?)`

	stmt, _ := r.db.Prepare(query)

	_, err := stmt.Exec(
		cpe.UUID,
		parameter.Name,
		parameter.Value.Value,
		parameter.Value.Type, //TODO: NORMALIZE
		"",                   //TODO: Flags support (R - Read, W - Write and more...)
		time.Now(),
		time.Now(),
	)

	if err != nil {
		fmt.Println(repository.ErrParameterCreating, err.Error())
		return false, err
	}

	return true, nil
}

func (r *MysqlCPERepositoryImpl) UpdateOrCreateParameter(cpe *cpe.CPE, parameter xml.ParameterValueStruct) (result bool, err error) {
	existParameter, err := r.FindParameter(cpe, parameter.Name)

	if existParameter == nil {
		fmt.Println("non exist param", existParameter)
		result, err = r.CreateParameter(cpe, parameter)
	} else {
		fmt.Println("param exist", existParameter)
		var query string = "UPDATE cpe_parameters SET value=?, type=?, flags=?, updated_at=? WHERE cpe_uuid=? and name = ?"
		stmt, _ := r.db.Prepare(query)

		_, err = stmt.Exec(
			parameter.Value.Value,
			parameter.Value.Type,
			"",
			time.Now(),
			cpe.UUID,
			parameter.Name,
		)

		if err != nil {
			fmt.Println("ERROR", err.Error())
			result = false
		}
	}

	return
}

func (r *MysqlCPERepositoryImpl) SaveParameters(cpe *cpe.CPE) (bool, error) {

	for _, parameterValue := range cpe.ParameterValues {
		fmt.Println("param value", parameterValue)
		_, err := r.UpdateOrCreateParameter(cpe, parameterValue)

		if err != nil {
			fmt.Println(repository.ErrParameterCreating, err.Error())
			return false, err
		}
	}

	return true, nil
}

func (r *MysqlCPERepositoryImpl) LoadParameters(cpe *cpe.CPE) (bool, error) {
	panic("implement me")
}
