package pgsql

// Create 单条插入
func (p *PGSQL) Create(obj *interface{}) error {
	return p.DB.Create(obj).Error
}

// CreateBatch 批量插入
func (p *PGSQL) CreateBatch(objs []*interface{}) error {
	return p.DB.Create(&objs).Error
}

// Save 更新或插入
func (p *PGSQL) Save(obj *interface{}) error {
	return p.DB.Save(obj).Error
}

// Update 指定字段更新
func (p *PGSQL) Update(obj *interface{}, fields map[string]interface{}) error {
	return p.DB.Model(obj).Updates(fields).Error
}

// UpdateBatch 批量更新
func (p *PGSQL) UpdateBatch(objs []*interface{}, fields []string) error {
	for _, obj := range objs {
		if err := p.DB.Model(obj).Select(fields).Updates(obj).Error; err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除
func (p *PGSQL) Delete(obj *interface{}, conds ...interface{}) error {
	return p.DB.Delete(obj, conds...).Error
}

// SoftDelete 软删除
func (p *PGSQL) SoftDelete(obj *interface{}, conds ...interface{}) error {
	return p.DB.Delete(obj, conds...).Error
}

// Restore 恢复软删除数据
func (p *PGSQL) Restore(obj *interface{}, cond map[string]interface{}) error {
	return p.DB.Unscoped().Model(obj).Where(cond).Update("deleted_at", nil).Error
}

// First 查询第一条
func (p *PGSQL) First(cond map[string]interface{}, out *interface{}) error {
	return p.DB.Where(cond).First(out).Error
}

// Find 查询列表
func (p *PGSQL) Find(cond map[string]interface{}, out *[]interface{}) error {
	return p.DB.Where(cond).Find(out).Error
}

// FindByIDs 批量主键查询
func (p *PGSQL) FindByIDs(ids []interface{}, out *[]interface{}) error {
	return p.DB.Where("id IN ?", ids).Find(out).Error
}

// Exists 是否存在
func (p *PGSQL) Exists(cond map[string]interface{}) (bool, error) {
	var count int64
	if err := p.DB.Model(new(interface{})).Where(cond).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
