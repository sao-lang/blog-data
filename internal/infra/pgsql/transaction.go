package pgsql

import "gorm.io/gorm"

// 事务执行器
func (p *PGSQL[T]) Transaction(fc func(tx *gorm.DB) error) error {
	return p.DB.Transaction(fc)
}

// 嵌套事务
func TransactionWithTx(tx *gorm.DB, fc func(tx *gorm.DB) error) error {
	return tx.Transaction(fc)
}
