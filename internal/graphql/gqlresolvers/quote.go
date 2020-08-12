package gqlresolvers

import (
	"context"

	"go-webapp-example/internal/graphql/gqlmodels"
	"go-webapp-example/internal/pkg/entity"
	"go-webapp-example/internal/pkg/quote"
	"go-webapp-example/pkg/db"
)

// Queries

func (r *queryResolver) Quotes(ctx context.Context) ([]*entity.Quote, error) {
	return r.Services.Quote.Get(ctx)
}
func (r *queryResolver) Quote(ctx context.Context, id int) (*entity.Quote, error) {
	return r.Services.Quote.Find(ctx, id)
}

// Mutations

func (r *mutationResolver) CreateQuote(ctx context.Context, input gqlmodels.QuoteInput) (*entity.Quote, error) {
	if err := quote.ValidateCreateRequest(&input); err.Failed() {
		return nil, addErrors(ctx, err)
	}
	return r.Services.Quote.Create(ctx, toQuoteEntity(input))
}

func (r *mutationResolver) UpdateQuote(ctx context.Context, input gqlmodels.QuoteInput) (*entity.Quote, error) {
	if err := quote.ValidateUpdateRequest(&input); err.Failed() {
		return nil, addErrors(ctx, err)
	}
	return r.Services.Quote.Update(ctx, toQuoteEntity(input))
}

func (r *mutationResolver) DeleteQuote(ctx context.Context, ids []int) ([]*entity.Quote, error) {
	tx, err := r.Services.DB.Begin()
	if err != nil {
		return nil, err
	}
	res, err := r.Services.Quote.Delete(ctx, tx, ids)
	if err != nil {
		return nil, db.RollbackError(tx, err)
	}
	return res, tx.Commit()
}

func toQuoteEntity(input gqlmodels.QuoteInput) *entity.Quote {
	return &entity.Quote{
		ID:      handleIntPtr(input.ID),
		Content: input.Content,
		Author:  input.Author,
	}
}
