package cluster

import (
	"context"
	"fmt"
	"io"
	"net/netip"
	"time"
)

type credentials struct {
	token     string
	clusterID string
}

type servers interface {
	checkHost(context.Context, host) error
	source(context.Context, host, netip.Addr) (credentials, error)
	inspect(context.Context, host, netip.Addr, credentials) (nodeState, error)
	join(context.Context, host, netip.Addr, credentials) error
	ready(context.Context, host, string) error
	verify(context.Context, config, credentials) error
}

// Enroll joins installed nodes. It never installs machines or resets cluster state.
func Enroll(ctx context.Context, environment string, progress io.Writer) error {
	cfg, err := loadConfig(environment)
	if err != nil {
		return err
	}
	remote, close, err := newSSHServers(ctx)
	if err != nil {
		return err
	}
	defer close()
	return enroll(ctx, cfg, remote, progress)
}

func enroll(ctx context.Context, cfg config, remote servers, progress io.Writer) error {
	for _, node := range append([]host{cfg.seed}, cfg.joiners...) {
		if err := remote.checkHost(ctx, node); err != nil {
			return err
		}
	}
	fmt.Fprintf(progress, "Waiting for %s and VIP %s...\n", cfg.seed.name, cfg.vip)
	var source credentials
	if err := wait(ctx, "initializer readiness", func() error {
		var err error
		source, err = remote.source(ctx, cfg.seed, cfg.vip)
		return err
	}); err != nil {
		return err
	}
	for _, node := range cfg.joiners {
		state, err := remote.inspect(ctx, node, cfg.vip, source)
		if err != nil {
			return fmt.Errorf("inspect %s: %w", node.name, err)
		}
		fmt.Fprintf(progress, "%s: %s\n", node.name, state)
	}
	for _, node := range cfg.joiners {
		current, err := remote.source(ctx, cfg.seed, cfg.vip)
		if err != nil {
			return err
		}
		if current != source {
			return fmt.Errorf("initializer identity or credentials changed; enrollment stopped")
		}
		if err := remote.join(ctx, node, cfg.vip, source); err != nil {
			return fmt.Errorf("enroll %s: %w; completed joins are preserved", node.name, err)
		}
		fmt.Fprintf(progress, "Waiting for %s to join...\n", node.name)
		if err := wait(ctx, node.name+" local API readiness", func() error {
			return remote.ready(ctx, node, source.clusterID)
		}); err != nil {
			return err
		}
	}
	fmt.Fprintln(progress, "Waiting for all expected control-plane nodes to become Ready...")
	return wait(ctx, "cluster membership", func() error {
		return remote.verify(ctx, cfg, source)
	})
}

func wait(ctx context.Context, description string, check func() error) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		err := check()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for %s: %w; last check: %v", description, ctx.Err(), err)
		case <-ticker.C:
		}
	}
}
