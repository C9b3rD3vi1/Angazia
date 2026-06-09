package database

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID      string
	Name    string
	Up      func(*gorm.DB) error
	Down    func(*gorm.DB) error
}

// RunMigrations runs all pending migrations
func RunMigrations() error {
	log.Println("🔄 Running manual migrations...")
	
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(); err != nil {
		return err
	}
	
	migrations := getAllMigrations()
	
	for _, migration := range migrations {
		// Check if migration already ran
		var count int64
		DB.Table("schema_migrations").Where("id = ?", migration.ID).Count(&count)
		
		if count == 0 {
			log.Printf("📝 Running migration: %s", migration.Name)
			
			// Run migration in transaction
			err := DB.Transaction(func(tx *gorm.DB) error {
				if err := migration.Up(tx); err != nil {
					return err
				}
				
				// Record migration
				if err := tx.Table("schema_migrations").Create(map[string]interface{}{
					"id":          migration.ID,
					"name":        migration.Name,
					"executed_at": gorm.Expr("NOW()"),
				}).Error; err != nil {
					return err
				}
				
				return nil
			})
			
			if err != nil {
				return fmt.Errorf("migration %s failed: %w", migration.ID, err)
			}
			
			log.Printf("✅ Migration completed: %s", migration.Name)
		}
	}
	
	log.Println("✅ All migrations completed successfully")
	return nil
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
func createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	return DB.Exec(query).Error
}

// getAllMigrations returns all available migrations
func getAllMigrations() []Migration {
	return []Migration{
		{
			ID:   "20250101000001",
			Name: "Add full-text search for jobs",
			Up: func(tx *gorm.DB) error {
				// Create full-text search vector column
				query := `
					ALTER TABLE jobs ADD COLUMN IF NOT EXISTS search_vector tsvector;
					UPDATE jobs SET search_vector = 
						setweight(to_tsvector('english', COALESCE(title, '')), 'A') ||
						setweight(to_tsvector('english', COALESCE(description, '')), 'B') ||
						setweight(to_tsvector('english', array_to_string(required_skills, ' ')), 'B');
					CREATE INDEX IF NOT EXISTS idx_jobs_search ON jobs USING GIN(search_vector);
				`
				return tx.Exec(query).Error
			},
			Down: func(tx *gorm.DB) error {
				query := `
					DROP INDEX IF EXISTS idx_jobs_search;
					ALTER TABLE jobs DROP COLUMN IF EXISTS search_vector;
				`
				return tx.Exec(query).Error
			},
		},
		{
			ID:   "20250101000002",
			Name: "Add triggers for updated_at timestamps",
			Up: func(tx *gorm.DB) error {
				fn := `
					CREATE OR REPLACE FUNCTION update_updated_at_column()
					RETURNS TRIGGER AS $$
					BEGIN
						NEW.updated_at = NOW();
						RETURN NEW;
					END;
					$$ language 'plpgsql';
				`
				if err := tx.Exec(fn).Error; err != nil {
					return err
				}
				
				tables := []string{"users", "employee_profiles", "employer_profiles", "jobs", "applications"}
				for _, table := range tables {
					trigger := fmt.Sprintf(`
						DROP TRIGGER IF EXISTS update_%s_updated_at ON %s;
						CREATE TRIGGER update_%s_updated_at
						BEFORE UPDATE ON %s
						FOR EACH ROW
						EXECUTE FUNCTION update_updated_at_column();
					`, table, table, table, table)
					
					if err := tx.Exec(trigger).Error; err != nil {
						log.Printf("Warning: Could not create trigger for %s: %v", table, err)
					}
				}
				
				return nil
			},
			Down: func(tx *gorm.DB) error {
				tables := []string{"users", "employee_profiles", "employer_profiles", "jobs", "applications"}
				for _, table := range tables {
					trigger := fmt.Sprintf("DROP TRIGGER IF EXISTS update_%s_updated_at ON %s", table, table)
					if err := tx.Exec(trigger).Error; err != nil {
						log.Printf("Warning: Could not drop trigger for %s: %v", table, err)
					}
				}
				return tx.Exec("DROP FUNCTION IF EXISTS update_updated_at_column()").Error
			},
		},
		{
			ID:   "20250101000003",
			Name: "Add job applications count trigger",
			Up: func(tx *gorm.DB) error {
				fn := `
					CREATE OR REPLACE FUNCTION update_job_applications_count()
					RETURNS TRIGGER AS $$
					BEGIN
						IF TG_OP = 'INSERT' THEN
							UPDATE jobs SET applications_count = applications_count + 1 
							WHERE id = NEW.job_id;
						ELSIF TG_OP = 'DELETE' THEN
							UPDATE jobs SET applications_count = applications_count - 1 
							WHERE id = OLD.job_id;
						END IF;
						RETURN NULL;
					END;
					$$ language 'plpgsql';
				`
				if err := tx.Exec(fn).Error; err != nil {
					return err
				}
				
				trigger := `
					DROP TRIGGER IF EXISTS update_job_applications_count ON applications;
					CREATE TRIGGER update_job_applications_count
					AFTER INSERT OR DELETE ON applications
					FOR EACH ROW
					EXECUTE FUNCTION update_job_applications_count();
				`
				return tx.Exec(trigger).Error
			},
			Down: func(tx *gorm.DB) error {
				tx.Exec("DROP TRIGGER IF EXISTS update_job_applications_count ON applications")
				return tx.Exec("DROP FUNCTION IF EXISTS update_job_applications_count()").Error
			},
		},
		{
			ID:   "20250101000004",
			Name: "Rename company_linked_in to company_linkedin in employer_profiles",
			Up: func(tx *gorm.DB) error {
				query := `
					DO $$
					BEGIN
						IF EXISTS (
							SELECT 1 FROM information_schema.columns
							WHERE table_name = 'employer_profiles' AND column_name = 'company_linked_in'
						) THEN
							ALTER TABLE employer_profiles RENAME COLUMN company_linked_in TO company_linkedin;
						END IF;
					END $$;
				`
				return tx.Exec(query).Error
			},
			Down: func(tx *gorm.DB) error {
				query := `
					DO $$
					BEGIN
						IF EXISTS (
							SELECT 1 FROM information_schema.columns
							WHERE table_name = 'employer_profiles' AND column_name = 'company_linkedin'
						) THEN
							ALTER TABLE employer_profiles RENAME COLUMN company_linkedin TO company_linked_in;
						END IF;
					END $$;
				`
				return tx.Exec(query).Error
			},
		},
		{
			ID:   "20250101000005",
			Name: "Add composite indexes for common queries",
			Up: func(tx *gorm.DB) error {
				indexes := []string{
					"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_jobs_employer_active ON jobs(employer_id, is_active)",
					"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_applications_job_status ON applications(job_id, status)",
					"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_matches_employee_score_expires ON matches(employee_id, overall_score, expires_at)",
					"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_employee_profile_available ON employee_profiles(is_available, is_visible)",
				}
				
				for _, idx := range indexes {
					if err := tx.Exec(idx).Error; err != nil {
						log.Printf("Warning: Could not create index: %v", err)
					}
				}
				return nil
			},
			Down: func(tx *gorm.DB) error {
				indexes := []string{
					"DROP INDEX IF EXISTS idx_jobs_employer_active",
					"DROP INDEX IF EXISTS idx_applications_job_status",
					"DROP INDEX IF EXISTS idx_matches_employee_score_expires",
					"DROP INDEX IF EXISTS idx_employee_profile_available",
				}
				for _, idx := range indexes {
					if err := tx.Exec(idx).Error; err != nil {
						log.Printf("Warning: Could not drop index: %v", err)
					}
				}
				return nil
			},
		},
	}
}

// RollbackLastMigration rolls back the last migration
func RollbackLastMigration() error {
	var lastMigration struct {
		ID string
	}
	
	err := DB.Table("schema_migrations").Order("executed_at DESC").Limit(1).Scan(&lastMigration).Error
	if err != nil {
		return err
	}
	
	if lastMigration.ID == "" {
		return fmt.Errorf("no migrations to rollback")
	}
	
	migrations := getAllMigrations()
	for _, migration := range migrations {
		if migration.ID == lastMigration.ID {
			log.Printf("🔙 Rolling back migration: %s", migration.Name)
			
			err := DB.Transaction(func(tx *gorm.DB) error {
				if err := migration.Down(tx); err != nil {
					return err
				}
				
				if err := tx.Table("schema_migrations").Where("id = ?", migration.ID).Delete(nil).Error; err != nil {
					return err
				}
				
				return nil
			})
			
			if err != nil {
				return err
			}
			
			log.Printf("✅ Rollback completed: %s", migration.Name)
			return nil
		}
	}
	
	return fmt.Errorf("migration %s not found", lastMigration.ID)
}

// MigrationStatus shows migration status
func MigrationStatus() error {
	type MigrationRecord struct {
		ID         string
		Name       string
		ExecutedAt string
	}
	
	var records []MigrationRecord
	if err := DB.Table("schema_migrations").Order("executed_at").Find(&records).Error; err != nil {
		return err
	}
	
	log.Println("\n📊 Migration Status:")
	log.Println(strings.Repeat("-", 80))
	log.Printf("%-20s %-40s %-20s", "ID", "Name", "Executed At")
	log.Println(strings.Repeat("-", 80))
	
	for _, record := range records {
		log.Printf("%-20s %-40s %-20s", record.ID, record.Name, record.ExecutedAt)
	}
	
	log.Println(strings.Repeat("-", 80))
	log.Printf("Total migrations executed: %d\n", len(records))
	
	return nil
}