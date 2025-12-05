// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"time"
)

// Node represents a node/server in the cluster.
type Node struct {
	ID         int64
	Name       string
	Host       string
	Port       int
	Protocol   string
	IsActive   bool
	LastSeenAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NodeRepository handles database operations for nodes.
type NodeRepository struct {
	db *sql.DB
}

// NewNodeRepository creates a new node repository.
func NewNodeRepository(db *sql.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// Create inserts a new node into the database.
func (r *NodeRepository) Create(node *Node) error {
	result, err := r.db.Exec(
		`INSERT INTO nodes (name, host, port, protocol, is_active, last_seen_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		node.Name,
		node.Host,
		node.Port,
		node.Protocol,
		node.IsActive,
		node.LastSeenAt,
	)
	if err != nil {
		return err
	}

	node.ID, err = result.LastInsertId()
	return err
}

// GetByID retrieves a node by ID.
func (r *NodeRepository) GetByID(id int64) (*Node, error) {
	node := &Node{}
	err := r.db.QueryRow(
		`SELECT id, name, host, port, protocol, is_active, last_seen_at, created_at, updated_at
		FROM nodes
		WHERE id = ?`,
		id,
	).Scan(
		&node.ID,
		&node.Name,
		&node.Host,
		&node.Port,
		&node.Protocol,
		&node.IsActive,
		&node.LastSeenAt,
		&node.CreatedAt,
		&node.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return node, nil
}

// GetByName retrieves a node by name.
func (r *NodeRepository) GetByName(name string) (*Node, error) {
	node := &Node{}
	err := r.db.QueryRow(
		`SELECT id, name, host, port, protocol, is_active, last_seen_at, created_at, updated_at
		FROM nodes
		WHERE name = ?`,
		name,
	).Scan(
		&node.ID,
		&node.Name,
		&node.Host,
		&node.Port,
		&node.Protocol,
		&node.IsActive,
		&node.LastSeenAt,
		&node.CreatedAt,
		&node.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return node, nil
}

// Update updates a node's information.
func (r *NodeRepository) Update(node *Node) error {
	_, err := r.db.Exec(
		`UPDATE nodes SET
			name = ?, host = ?, port = ?, protocol = ?, is_active = ?, last_seen_at = ?, updated_at = ?
		WHERE id = ?`,
		node.Name,
		node.Host,
		node.Port,
		node.Protocol,
		node.IsActive,
		node.LastSeenAt,
		time.Now().UTC(),
		node.ID,
	)
	return err
}

// Delete removes a node from the database.
func (r *NodeRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM nodes WHERE id = ?", id)
	return err
}

// List retrieves all nodes.
func (r *NodeRepository) List() ([]*Node, error) {
	rows, err := r.db.Query(
		`SELECT id, name, host, port, protocol, is_active, last_seen_at, created_at, updated_at
		FROM nodes
		ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
		if err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.Host,
			&node.Port,
			&node.Protocol,
			&node.IsActive,
			&node.LastSeenAt,
			&node.CreatedAt,
			&node.UpdatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}

// ListActive retrieves all active nodes.
func (r *NodeRepository) ListActive() ([]*Node, error) {
	rows, err := r.db.Query(
		`SELECT id, name, host, port, protocol, is_active, last_seen_at, created_at, updated_at
		FROM nodes
		WHERE is_active = 1
		ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
		if err := rows.Scan(
			&node.ID,
			&node.Name,
			&node.Host,
			&node.Port,
			&node.Protocol,
			&node.IsActive,
			&node.LastSeenAt,
			&node.CreatedAt,
			&node.UpdatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}

// UpdateLastSeen updates the node's last seen timestamp.
func (r *NodeRepository) UpdateLastSeen(id int64) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE nodes SET
			last_seen_at = ?, updated_at = ?
		WHERE id = ?`,
		now,
		now,
		id,
	)
	return err
}

// NodeMeta represents metadata associated with a node.
type NodeMeta struct {
	ID        int64
	Key       string
	Value     string
	NodeID    int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NodeMetaRepository handles database operations for node metadata.
type NodeMetaRepository struct {
	db *sql.DB
}

// NewNodeMetaRepository creates a new node meta repository.
func NewNodeMetaRepository(db *sql.DB) *NodeMetaRepository {
	return &NodeMetaRepository{db: db}
}

// Create inserts new metadata for a node.
func (r *NodeMetaRepository) Create(nodeID int64, key, value string) error {
	_, err := r.db.Exec(
		"INSERT INTO nodes_meta (node_id, key, value) VALUES (?, ?, ?)",
		nodeID,
		key,
		value,
	)
	return err
}

// Get retrieves metadata for a node by key.
func (r *NodeMetaRepository) Get(nodeID int64, key string) (*NodeMeta, error) {
	meta := &NodeMeta{}
	err := r.db.QueryRow(
		`SELECT id, key, value, node_id, created_at, updated_at
		FROM nodes_meta
		WHERE node_id = ? AND key = ?`,
		nodeID,
		key,
	).Scan(
		&meta.ID,
		&meta.Key,
		&meta.Value,
		&meta.NodeID,
		&meta.CreatedAt,
		&meta.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// Update updates metadata for a node.
func (r *NodeMetaRepository) Update(nodeID int64, key, value string) error {
	_, err := r.db.Exec(
		`UPDATE nodes_meta SET
			value = ?, updated_at = ?
		WHERE node_id = ? AND key = ?`,
		value,
		time.Now().UTC(),
		nodeID,
		key,
	)
	return err
}

// Delete removes metadata for a node.
func (r *NodeMetaRepository) Delete(nodeID int64, key string) error {
	_, err := r.db.Exec(
		"DELETE FROM nodes_meta WHERE node_id = ? AND key = ?",
		nodeID,
		key,
	)
	return err
}

// ListByNode retrieves all metadata for a node.
func (r *NodeMetaRepository) ListByNode(nodeID int64) ([]*NodeMeta, error) {
	rows, err := r.db.Query(
		`SELECT id, key, value, node_id, created_at, updated_at
		FROM nodes_meta
		WHERE node_id = ?
		ORDER BY key`,
		nodeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metadata []*NodeMeta
	for rows.Next() {
		meta := &NodeMeta{}
		if err := rows.Scan(
			&meta.ID,
			&meta.Key,
			&meta.Value,
			&meta.NodeID,
			&meta.CreatedAt,
			&meta.UpdatedAt,
		); err != nil {
			return nil, err
		}
		metadata = append(metadata, meta)
	}

	return metadata, rows.Err()
}

// Upsert inserts or updates metadata for a node.
func (r *NodeMetaRepository) Upsert(nodeID int64, key, value string) error {
	existing, err := r.Get(nodeID, key)
	if err != nil {
		return err
	}

	if existing == nil {
		return r.Create(nodeID, key, value)
	}

	return r.Update(nodeID, key, value)
}
