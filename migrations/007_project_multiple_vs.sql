-- Migration 007: Support multiple VS identifiers per project
-- This allows projects to accept payments with different/mistyped VS

-- Create project_vs table for multiple VS per project
CREATE TABLE IF NOT EXISTS project_vs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    vs TEXT NOT NULL,
    note TEXT,  -- Optional note like "original" or "typo by XY"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, vs)
);

-- Migrate existing VS from projects table to project_vs
INSERT OR IGNORE INTO project_vs (project_id, vs, note)
SELECT id, payments_id, 'primary' FROM projects WHERE payments_id IS NOT NULL AND payments_id != '';

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_project_vs_vs ON project_vs(vs);
CREATE INDEX IF NOT EXISTS idx_project_vs_project ON project_vs(project_id);
