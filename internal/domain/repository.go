package domain

import (
	"database/sql"
	"fmt"
	
	"ipup-go/pkg/types"
)

// Repository 域名数据访问层
type Repository struct {
	db *sql.DB
}

// NewRepository 创建域名仓库
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建域名配置
func (r *Repository) Create(domain types.Domain) (int64, error) {
	insertSQL := `
	INSERT INTO domains (domain, provider, token, access_key_id, access_key_secret, interval, enabled, created_at, updated_at) 
	VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	
	result, err := r.db.Exec(insertSQL, domain.Domain, domain.Provider, domain.Token, domain.AccessKeyID, domain.AccessKeySecret, domain.Interval, domain.Enabled)
	if err != nil {
		return 0, fmt.Errorf("插入域名失败：%w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取插入 ID 失败：%w", err)
	}
	
	return id, nil
}

// Update 更新域名配置
func (r *Repository) Update(domain types.Domain) error {
	updateSQL := `
	UPDATE domains 
	SET domain = ?, provider = ?, token = ?, access_key_id = ?, access_key_secret = ?, interval = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP 
	WHERE id = ?
	`
	
	_, err := r.db.Exec(updateSQL, domain.Domain, domain.Provider, domain.Token, domain.AccessKeyID, domain.AccessKeySecret, domain.Interval, domain.Enabled, domain.ID)
	if err != nil {
		return fmt.Errorf("更新域名失败：%w", err)
	}
	
	return nil
}

// Delete 删除域名配置
func (r *Repository) Delete(id int) error {
	deleteSQL := `DELETE FROM domains WHERE id = ?`
	_, err := r.db.Exec(deleteSQL, id)
	if err != nil {
		return fmt.Errorf("删除域名失败：%w", err)
	}
	
	return nil
}

// GetByID 根据 ID 获取域名
func (r *Repository) GetByID(id int) (types.Domain, error) {
	querySQL := `
	SELECT id, domain, provider, token, 
	       COALESCE(access_key_id, ''), COALESCE(access_key_secret, ''),
	       interval, enabled, 
	       COALESCE(current_ip, ''), COALESCE(last_update, ''), 
	       COALESCE(created_at, ''), COALESCE(updated_at, '')
	FROM domains 
	WHERE id = ?
	`
	
	var d types.Domain
	err := r.db.QueryRow(querySQL, id).Scan(
		&d.ID, &d.Domain, &d.Provider, &d.Token, &d.AccessKeyID, &d.AccessKeySecret,
		&d.Interval, &d.Enabled,
		&d.CurrentIP, &d.LastUpdate, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return types.Domain{}, fmt.Errorf("域名不存在")
	}
	if err != nil {
		return types.Domain{}, fmt.Errorf("查询域名失败：%w", err)
	}
	
	return d, nil
}

// GetByDomain 根据域名获取配置
func (r *Repository) GetByDomain(domain string) (types.Domain, error) {
	querySQL := `
	SELECT id, domain, provider, token, 
	       COALESCE(access_key_id, ''), COALESCE(access_key_secret, ''),
	       interval, enabled, 
	       COALESCE(current_ip, ''), COALESCE(last_update, ''), 
	       COALESCE(created_at, ''), COALESCE(updated_at, '')
	FROM domains 
	WHERE domain = ?
	`
	
	var d types.Domain
	err := r.db.QueryRow(querySQL, domain).Scan(
		&d.ID, &d.Domain, &d.Provider, &d.Token, &d.AccessKeyID, &d.AccessKeySecret,
		&d.Interval, &d.Enabled,
		&d.CurrentIP, &d.LastUpdate, &d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return types.Domain{}, fmt.Errorf("域名不存在")
	}
	if err != nil {
		return types.Domain{}, fmt.Errorf("查询域名失败：%w", err)
	}
	
	return d, nil
}

// List 获取所有域名列表
func (r *Repository) List() ([]types.Domain, error) {
	querySQL := `
	SELECT id, domain, provider, token, 
	       COALESCE(access_key_id, ''), COALESCE(access_key_secret, ''),
	       interval, enabled, 
	       COALESCE(current_ip, ''), COALESCE(last_update, ''), 
	       COALESCE(created_at, ''), COALESCE(updated_at, '')
	FROM domains 
	ORDER BY domain
	`
	
	rows, err := r.db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("查询域名列表失败：%w", err)
	}
	defer rows.Close()
	
	var domains []types.Domain
	for rows.Next() {
		var d types.Domain
		if err := rows.Scan(
			&d.ID, &d.Domain, &d.Provider, &d.Token, &d.AccessKeyID, &d.AccessKeySecret,
			&d.Interval, &d.Enabled,
			&d.CurrentIP, &d.LastUpdate, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			continue
		}
		domains = append(domains, d)
	}
	
	return domains, nil
}

// ListEnabled 获取所有启用的域名
func (r *Repository) ListEnabled() ([]types.Domain, error) {
	querySQL := `
	SELECT id, domain, provider, token, 
	       COALESCE(access_key_id, ''), COALESCE(access_key_secret, ''),
	       interval, enabled, 
	       COALESCE(current_ip, ''), COALESCE(last_update, ''), 
	       COALESCE(created_at, ''), COALESCE(updated_at, '')
	FROM domains 
	WHERE enabled = 1
	ORDER BY domain
	`
	
	rows, err := r.db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("查询启用域名失败：%w", err)
	}
	defer rows.Close()
	
	var domains []types.Domain
	for rows.Next() {
		var d types.Domain
		if err := rows.Scan(
			&d.ID, &d.Domain, &d.Provider, &d.Token, &d.AccessKeyID, &d.AccessKeySecret,
			&d.Interval, &d.Enabled,
			&d.CurrentIP, &d.LastUpdate, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			continue
		}
		domains = append(domains, d)
	}
	
	return domains, nil
}

// Toggle 切换域名启用状态
func (r *Repository) Toggle(id int) (bool, error) {
	// 先获取当前状态
	d, err := r.GetByID(id)
	if err != nil {
		return false, err
	}
	
	newEnabled := !d.Enabled
	updateSQL := `UPDATE domains SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = r.db.Exec(updateSQL, newEnabled, id)
	if err != nil {
		return false, fmt.Errorf("切换状态失败：%w", err)
	}
	
	return newEnabled, nil
}

// UpdateIP 更新域名的 IP 地址
func (r *Repository) UpdateIP(id int, ip string) error {
	updateSQL := `
	UPDATE domains 
	SET current_ip = ?, last_update = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
	WHERE id = ?
	`
	
	_, err := r.db.Exec(updateSQL, ip, id)
	if err != nil {
		return fmt.Errorf("更新 IP 失败：%w", err)
	}
	
	return nil
}

