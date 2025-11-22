package pgsql

import "gorm.io/gorm"

// Upsert 冲突更新
func (p *PGSQL) Upsert[T any](obj *T, conflictFields []string) error {
    return p.DB.Clauses(
        gorm.Clauses{gorm.OnConflict{
            Columns:   conflictFields,
            DoUpdates: gorm.AssignmentColumns([]string{}),
        }},
    ).Create(obj).Error
}

// 原生 SQL 查询
func (p *PGSQL) RawQuery(dest interface{}, query string, args ...interface{}) error {
    return p.DB.Raw(query, args...).Scan(dest).Error
}

// 原生 SQL 执行
func (p *PGSQL) ExecSQL(query string, args ...interface{}) error {
    return p.DB.Exec(query, args...).Error
}
