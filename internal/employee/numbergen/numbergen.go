package numbergen

import (
	"context"
	"fmt"
)

type SequenceRepository interface {
	GetCurrent(ctx context.Context, designationCode string) (*Sequence, error)
	GetForUpdate(ctx context.Context, designationCode string) (*Sequence, error)
	CreateSequence(ctx context.Context, designationCode, prefix string, lastSequence int) error
	Increment(ctx context.Context, designationCode string, nextSequence int) error
	SetMinimumSequence(ctx context.Context, designationCode string, minSequence int) error
}

type Sequence struct {
	DesignationCode string
	Prefix          string
	LastSequence    int
}

type Generator struct {
	seqRepo SequenceRepository
	prefix  string
}

func New(seqRepo SequenceRepository, prefix string) *Generator {
	return &Generator{seqRepo: seqRepo, prefix: prefix}
}

func (g *Generator) Peek(ctx context.Context, designationCode string) (string, error) {
	seq, err := g.seqRepo.GetCurrent(ctx, designationCode)
	if err != nil {
		return "", fmt.Errorf("peek: %w", err)
	}
	if seq == nil {
		return g.formatNumber(designationCode, 1), nil
	}
	return g.formatNumber(designationCode, seq.LastSequence+1), nil
}

func (g *Generator) Generate(ctx context.Context, designationCode string) (string, error) {
	seq, err := g.seqRepo.GetForUpdate(ctx, designationCode)
	if err != nil {
		return "", fmt.Errorf("generate: %w", err)
	}

	if seq == nil {
		if err := g.seqRepo.CreateSequence(ctx, designationCode, g.prefix, 1); err != nil {
			seq, err = g.seqRepo.GetForUpdate(ctx, designationCode)
			if err != nil || seq == nil {
				return "", fmt.Errorf("generate: %w", err)
			}
			next := seq.LastSequence + 1
			if err := g.seqRepo.Increment(ctx, designationCode, next); err != nil {
				return "", fmt.Errorf("generate increment: %w", err)
			}
			return g.formatNumber(designationCode, next), nil
		}
		return g.formatNumber(designationCode, 1), nil
	}

	next := seq.LastSequence + 1
	if err := g.seqRepo.Increment(ctx, designationCode, next); err != nil {
		return "", fmt.Errorf("generate increment: %w", err)
	}
	return g.formatNumber(designationCode, next), nil
}

func (g *Generator) formatNumber(code string, seq int) string {
	return fmt.Sprintf("%s%s%02d", g.prefix, code, seq)
}

func (g *Generator) EnsureAtLeast(ctx context.Context, designationCode string, minSeq int) error {
	seq, err := g.seqRepo.GetForUpdate(ctx, designationCode)
	if err != nil {
		return err
	}
	if seq == nil {
		return g.seqRepo.CreateSequence(ctx, designationCode, g.prefix, minSeq)
	}
	if seq.LastSequence >= minSeq {
		return nil
	}
	return g.seqRepo.SetMinimumSequence(ctx, designationCode, minSeq)
}
