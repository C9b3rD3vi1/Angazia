package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/google/uuid"
	
	"github.com/C9b3rD3vi1/Angazia/internal/models"
)

type PaymentRepository interface {
	// Payments
	CreatePayment(ctx context.Context, payment *models.Payment) error
	GetPayment(ctx context.Context, id string) (*models.Payment, error)
	GetPaymentByReference(ctx context.Context, reference string) (*models.Payment, error)
	GetPaymentByTransactionID(ctx context.Context, transactionID string) (*models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, id, status string, transactionID string, paidAt *time.Time) error
	UpdatePaymentReceipt(ctx context.Context, id, receiptURL string) error
	ListUserPayments(ctx context.Context, userID string, page, limit int) ([]*models.Payment, int64, error)
	
	// Payment Intents
	CreatePaymentIntent(ctx context.Context, intent *models.PaymentIntent) error
	GetPaymentIntent(ctx context.Context, id string) (*models.PaymentIntent, error)
	GetPaymentIntentByInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentIntent, error)
	UpdatePaymentIntentStatus(ctx context.Context, id, status string) error
	DeleteExpiredIntents(ctx context.Context) error
	
	// Invoices
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error
	GetInvoice(ctx context.Context, id string) (*models.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, number string) (*models.Invoice, error)
	GetInvoiceByPaymentID(ctx context.Context, paymentID string) (*models.Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, id, status string, paidAt *time.Time) error
	UpdateInvoicePDF(ctx context.Context, id, pdfURL string) error
	ListUserInvoices(ctx context.Context, userID string, page, limit int) ([]*models.Invoice, int64, error)
	
	// Invoice Items
	CreateInvoiceItem(ctx context.Context, item *models.InvoiceItem) error
	GetInvoiceItems(ctx context.Context, invoiceID string) ([]*models.InvoiceItem, error)
}

type PaymentRepositoryImpl struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &PaymentRepositoryImpl{db: db}
}

func (r *PaymentRepositoryImpl) CreatePayment(ctx context.Context, payment *models.Payment) error {
	payment.ID = uuid.New().String()
	payment.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *PaymentRepositoryImpl) GetPayment(ctx context.Context, id string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Subscription").
		Where("id = ?", id).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepositoryImpl) GetPaymentByReference(ctx context.Context, reference string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Where("reference = ?", reference).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepositoryImpl) GetPaymentByTransactionID(ctx context.Context, transactionID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Where("transaction_id = ?", transactionID).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepositoryImpl) UpdatePaymentStatus(ctx context.Context, id, status string, transactionID string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
		"updated_at": time.Now(),
	}
	if transactionID != "" {
		updates["transaction_id"] = transactionID
	}
	if paidAt != nil {
		updates["paid_at"] = paidAt
	}
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *PaymentRepositoryImpl) UpdatePaymentReceipt(ctx context.Context, id, receiptURL string) error {
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("id = ?", id).
		Update("receipt_url", receiptURL).Error
}

func (r *PaymentRepositoryImpl) ListUserPayments(ctx context.Context, userID string, page, limit int) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("user_id = ?", userID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Subscription").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&payments).Error
	
	return payments, total, err
}

func (r *PaymentRepositoryImpl) CreatePaymentIntent(ctx context.Context, intent *models.PaymentIntent) error {
	intent.ID = uuid.New().String()
	intent.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(intent).Error
}

func (r *PaymentRepositoryImpl) GetPaymentIntent(ctx context.Context, id string) (*models.PaymentIntent, error) {
	var intent models.PaymentIntent
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&intent).Error
	if err != nil {
		return nil, err
	}
	return &intent, nil
}

func (r *PaymentRepositoryImpl) GetPaymentIntentByInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentIntent, error) {
	var intent models.PaymentIntent
	err := r.db.WithContext(ctx).Where("invoice_id = ?", invoiceID).First(&intent).Error
	if err != nil {
		return nil, err
	}
	return &intent, nil
}

func (r *PaymentRepositoryImpl) UpdatePaymentIntentStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.PaymentIntent{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *PaymentRepositoryImpl) DeleteExpiredIntents(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", time.Now(), "pending").
		Delete(&models.PaymentIntent{}).Error
}

func (r *PaymentRepositoryImpl) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	invoice.ID = uuid.New().String()
	invoice.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(invoice).Error
}

func (r *PaymentRepositoryImpl) GetInvoice(ctx context.Context, id string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Subscription").
		Preload("Payment").
		Where("id = ?", id).
		First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *PaymentRepositoryImpl) GetInvoiceByNumber(ctx context.Context, number string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.WithContext(ctx).Where("invoice_number = ?", number).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *PaymentRepositoryImpl) GetInvoiceByPaymentID(ctx context.Context, paymentID string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *PaymentRepositoryImpl) UpdateInvoiceStatus(ctx context.Context, id, status string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
		"updated_at": time.Now(),
	}
	if paidAt != nil {
		updates["paid_at"] = paidAt
	}
	return r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *PaymentRepositoryImpl) UpdateInvoicePDF(ctx context.Context, id, pdfURL string) error {
	return r.db.WithContext(ctx).
		Model(&models.Invoice{}).
		Where("id = ?", id).
		Update("pdf_url", pdfURL).Error
}

func (r *PaymentRepositoryImpl) ListUserInvoices(ctx context.Context, userID string, page, limit int) ([]*models.Invoice, int64, error) {
	var invoices []*models.Invoice
	var total int64
	
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Where("user_id = ?", userID)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	offset := (page - 1) * limit
	err := query.
		Preload("Subscription").
		Preload("Payment").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&invoices).Error
	
	return invoices, total, err
}

func (r *PaymentRepositoryImpl) CreateInvoiceItem(ctx context.Context, item *models.InvoiceItem) error {
	item.ID = uuid.New().String()
	item.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *PaymentRepositoryImpl) GetInvoiceItems(ctx context.Context, invoiceID string) ([]*models.InvoiceItem, error) {
	var items []*models.InvoiceItem
	err := r.db.WithContext(ctx).
		Where("invoice_id = ?", invoiceID).
		Find(&items).Error
	return items, err
}