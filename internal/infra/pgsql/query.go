package pgsql

// ---------------- 链式查询方法 ----------------

// 条件查询 map
func (p *PGSQL[T]) WhereMap(conds map[string]interface{}) *PGSQL[T] {
	p.DB = p.DB.Where(conds)
	return p
}

// 条件查询结构体
func (p *PGSQL[T]) WhereStruct(cond T) *PGSQL[T] {
	p.DB = p.DB.Where(&cond)
	return p
}

// 范围查询
func (p *PGSQL[T]) WhereRange(field string, min, max interface{}) *PGSQL[T] {
	if min != nil && max != nil {
		p.DB = p.DB.Where(field+" BETWEEN ? AND ?", min, max)
	} else if min != nil {
		p.DB = p.DB.Where(field+" >= ?", min)
	} else if max != nil {
		p.DB = p.DB.Where(field+" <= ?", max)
	}
	return p
}

// 模糊查询
func (p *PGSQL[T]) Like(field string, value string) *PGSQL[T] {
	p.DB = p.DB.Where(field+" ILIKE ?", "%"+value+"%")
	return p
}

// 批量 IN 查询
func (p *PGSQL[T]) FindIn(field string, values []interface{}) *PGSQL[T] {
	p.DB = p.DB.Where(field+" IN ?", values)
	return p
}

// 排序
func (p *PGSQL[T]) Order(field string, desc bool) *PGSQL[T] {
	order := field
	if desc {
		order += " DESC"
	} else {
		order += " ASC"
	}
	p.DB = p.DB.Order(order)
	return p
}

// 分页
func (p *PGSQL[T]) Paginate(page, pageSize int) *PGSQL[T] {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	p.DB = p.DB.Offset(offset).Limit(pageSize)
	return p
}

// ---------------- 查询方法 ----------------

// 分页 + 条件查询
func (p *PGSQL[T]) FindWithPage(page, pageSize int) (*PageResult[T], error) {
	var list []T
	var total int64

	// 统计总数
	if err := p.DB.Model(new(T)).Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := p.DB.Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, err
	}

	pageCount := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &PageResult[T]{
		List:      list,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		PageCount: pageCount,
	}, nil
}

// 查询全部
func (p *PGSQL[T]) FindAll() ([]T, error) {
	var list []T
	if err := p.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
