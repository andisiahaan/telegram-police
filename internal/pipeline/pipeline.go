package pipeline

import (
	"github.com/andisiahaan/telegram-police/internal/action"
	"github.com/andisiahaan/telegram-police/internal/botctx"
	"github.com/andisiahaan/telegram-police/internal/config"
	"github.com/andisiahaan/telegram-police/internal/filter"
	"github.com/andisiahaan/telegram-police/internal/gate"
	"github.com/andisiahaan/telegram-police/internal/ratelimit"
)

// Pipeline wires the processing stages together and runs them in order.
type Pipeline struct {
	gate         *gate.Gate
	floodChecker *ratelimit.FloodChecker
	filters      []filter.Filter
	resolver     *action.Resolver
	cfg          *config.Config
}

// New creates a Pipeline with all its dependencies injected.
func New(
	g *gate.Gate,
	flood *ratelimit.FloodChecker,
	filters []filter.Filter,
	resolver *action.Resolver,
	cfg *config.Config,
) *Pipeline {
	return &Pipeline{
		gate:         g,
		floodChecker: flood,
		filters:      filters,
		resolver:     resolver,
		cfg:          cfg,
	}
}

// Run processes a single message through the pipeline.
// Order: Gate → Flood → FilterChain → Resolver.
func (p *Pipeline) Run(ctx *botctx.Context) {
	// Skip everything for admins and excluded senders.
	if p.gate.ShouldBypass(ctx) {
		return
	}

	// Flood check — bail out early if triggered.
	if p.floodChecker.Check(ctx) {
		p.resolver.ResolveFlood(ctx)
		return
	}

	// Run all enabled filters and collect violations.
	filter.RunChain(p.filters, ctx)

	// Act on whatever violations were found.
	p.resolver.Resolve(ctx)
}
