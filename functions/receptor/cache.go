package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/pkg/errors"
)

type AlertMap struct {
	table dynamo.Table
}

func NewAlertMap(tableName, region string) *AlertMap {
	alertMap := AlertMap{}

	db := dynamo.New(session.New(), &aws.Config{Region: aws.String(region)})
	alertMap.table = db.Table(tableName)

	return &alertMap
}

type AlertRecord struct {
	AlertKey string       `json:"alert_key"`
	AlertID  string       `json:"alert_id"`
	Rule     string       `json:"rule"`
	ReportID lib.ReportID `json:"report_id"`
}

func GenAlertKey(alertID, rule string) string {
	data := fmt.Sprintf("%s=====%s", alertID, rule)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func (x *AlertMap) Lookup(alertID, rule string) (*lib.ReportID, error) {
	alertKey := GenAlertKey(alertID, rule)

	record := AlertRecord{}
	err := x.table.Get("alert_key", alertKey).One(&record)
	if err == dynamo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "Fail to get alert")
	}

	return &record.ReportID, nil
}

func (x *AlertMap) Create(alertID, rule string) (*lib.ReportID, error) {
	alertKey := GenAlertKey(alertID, rule)

	record := AlertRecord{
		AlertKey: alertKey,
		AlertID:  alertID,
		Rule:     rule,
		ReportID: lib.NewReportID(),
	}

	err := x.table.Put(&record).Run()
	if err != nil {
		return nil, errors.Wrap(err, "Fail to put alert map")
	}

	return &record.ReportID, nil
}