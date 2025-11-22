package pgsql

import "gorm.io/gorm"

// 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if page <= 0 {
            page = 1
        }
        if pageSize <= 0 {
            pageSize = 10
        }
        offset := (page - 1) * pageSize
        return db.Offset(offset).Limit(pageSize)
    }
}

// 模糊查询
func Like(field string, value string) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where(field+" ILIKE ?", "%"+value+"%")
    }
}

// 批量 IN 查询
func FindIn(field string, values []interface{}) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where(field+" IN ?", values)
    }
}

// 排序
func Order(field string, desc bool) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        order := field
        if desc {
            order += " DESC"
        } else {
            order += " ASC"
        }
        return db.Order(order)
    }
}

// 分页 + 条件查询
func (p *PGSQL) FindWithPage[T any](page, pageSize int, conds ...func(*gorm.DB) *gorm.DB) (*PageResult[T], error) {
    var list []T
    db := p.DB.Scopes(conds...)
    var total int64
    if err := db.Model(new(T)).Count(&total).Error; err != nil {
        return nil, err
    }
    if err := db.Scopes(Paginate(page, pageSize)).Find(&list).Error; err != nil {
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

// Map 条件查询
func WhereMap(conds map[string]interface{}) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where(conds)
    }
}

// 范围查询
func WhereRange(field string, min, max interface{}) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if min != nil && max != nil {
            return db.Where(field+" BETWEEN ? AND ?", min, max)
        }
        if min != nil {
            return db.Where(field+" >= ?", min)
        }
        if max != nil {
            return db.Where(field+" <= ?", max)
        }
        return db
    }
}
