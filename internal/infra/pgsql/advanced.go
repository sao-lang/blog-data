package pgsql

import "gorm.io/gorm/clause"

// Upsert 冲突更新
func (p *PGSQL) Upsert(obj *interface{}, conflictFields []string) error {
	// 将 string 转成 clause.Column
	columns := make([]clause.Column, len(conflictFields))
	for i, c := range conflictFields {
		columns[i] = clause.Column{Name: c}
	}

	return p.DB.Clauses(clause.OnConflict{
		Columns:   columns,
		DoUpdates: clause.AssignmentColumns(conflictFields), // 冲突字段自动更新自身
	}).Create(obj).Error
}

// 原生 SQL 查询
func (p *PGSQL) RawQuery(dest interface{}, query string, args ...interface{}) error {
	return p.DB.Raw(query, args...).Scan(dest).Error
}

// 原生 SQL 执行
func (p *PGSQL) ExecSQL(query string, args ...interface{}) error {
	return p.DB.Exec(query, args...).Error
}
