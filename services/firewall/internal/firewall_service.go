package internal

import (
	"context"

	"sync"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/firewall"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FirewallAction 防火墙动作
type FirewallAction string

const (
	Allow FirewallAction = "allow"
	Deny  FirewallAction = "deny"
)

// FirewallRule 防火墙规则
type FirewallRule struct {
	IP     string
	Action FirewallAction
}

// FirewallConfig 防火墙配置
type FirewallConfig struct {
	DefaultAction FirewallAction
	Rules         []FirewallRule
}

// FirewallService 防火墙服务实现
type FirewallService struct {
	firewall.UnimplementedFirewallServiceServer
	common.UnimplementedHealthCheckServiceServer

	config FirewallConfig
	mu     sync.RWMutex
}

// NewFirewallService 创建防火墙服务实例
func NewFirewallService() *FirewallService {
	return &FirewallService{}
}

// ConfigureFirewallRules 配置防火墙规则
func (s *FirewallService) ConfigureFirewallRules(config FirewallConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
}

// CheckAccess 检查访问权限
func (s *FirewallService) CheckAccess(ctx context.Context, req *firewall.CheckAccessRequest) (*firewall.CheckAccessResponse, error) {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	// 默认动作
	action := config.DefaultAction

	// 检查规则
	for _, rule := range config.Rules {
		if rule.IP == req.Ip {
			action = rule.Action
			break
		}
	}

	return &firewall.CheckAccessResponse{
			Allowed: action == Allow,
			Action:  string(action),
		},
		nil
}

// UpdateRule 更新防火墙规则
func (s *FirewallService) UpdateRule(ctx context.Context, req *firewall.FirewallRule) (*common.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查规则是否存在
	for i, rule := range s.config.Rules {
		if rule.IP == req.Ip {
			// 更新规则
			s.config.Rules[i] = FirewallRule{
				IP:     req.Ip,
				Action: FirewallAction(req.Action),
			}
			return &common.Response{
					Success: true,
					Message: "Rule updated successfully",
					Code:    200,
				},
				nil
		}
	}

	// 添加新规则
	s.config.Rules = append(s.config.Rules, FirewallRule{
		IP:     req.Ip,
		Action: FirewallAction(req.Action),
	})

	return &common.Response{
			Success: true,
			Message: "Rule added successfully",
			Code:    200,
		},
		nil
}

// DeleteRule 删除防火墙规则
func (s *FirewallService) DeleteRule(ctx context.Context, req *firewall.DeleteRuleRequest) (*common.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并删除规则
	for i, rule := range s.config.Rules {
		if rule.IP == req.Ip {
			s.config.Rules = append(s.config.Rules[:i], s.config.Rules[i+1:]...)
			return &common.Response{
					Success: true,
					Message: "Rule deleted successfully",
					Code:    200,
				},
				nil
		}
	}

	return &common.Response{
			Success: true,
			Message: "Rule not found",
			Code:    200,
		},
		nil
}

// GetRules 获取防火墙规则
func (s *FirewallService) GetRules(ctx context.Context, req *firewall.GetRulesRequest) (*firewall.GetRulesResponse, error) {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	// 转换为proto格式
	var rules []*firewall.FirewallRule
	for _, rule := range config.Rules {
		rules = append(rules, &firewall.FirewallRule{
			Ip:     rule.IP,
			Action: string(rule.Action),
		})
	}

	return &firewall.GetRulesResponse{
			Rules:         rules,
			DefaultAction: string(config.DefaultAction),
		},
		nil
}

// Check 健康检查
func (s *FirewallService) Check(ctx context.Context, req *common.HealthCheckRequest) (*common.HealthCheckResponse, error) {
	return &common.HealthCheckResponse{
			Status: common.HealthCheckResponse_SERVING,
		},
		nil
}

// Watch 健康检查监听
func (s *FirewallService) Watch(req *common.HealthCheckRequest, stream common.HealthCheckService_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
