package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/pkg/logger"
)

// ProxyService handles proxy configuration and management
type ProxyService struct {
	repos  *repository.Repositories
	logger *logger.Logger
}

// NewProxyService creates a new ProxyService
func NewProxyService(repos *repository.Repositories, logger *logger.Logger) *ProxyService {
	return &ProxyService{
		repos:  repos,
		logger: logger,
	}
}

// GetProxyConfig retrieves the current proxy configuration
func (s *ProxyService) GetProxyConfig() (*models.ProxyConfig, error) {
	return s.repos.Proxy.GetProxyConfig()
}

// UpdateProxyConfig updates the proxy configuration
func (s *ProxyService) UpdateProxyConfig(config *models.ProxyConfig) error {
	return s.repos.Proxy.UpdateProxyConfig(config)
}

// AddDomain adds a new domain with validation
func (s *ProxyService) AddDomain(name string, isPrimary bool) (*models.Domain, error) {
	// Validate domain format
	if err := s.validateDomainName(name); err != nil {
		return nil, err
	}

	domain := &models.Domain{
		Name:      strings.ToLower(strings.TrimSpace(name)),
		IsPrimary: isPrimary,
	}

	if err := s.repos.Proxy.AddDomain(domain); err != nil {
		return nil, err
	}

	s.logger.Info("Domain added", "domain", name, "primary", isPrimary)
	return domain, nil
}

// GetDomains retrieves all configured domains
func (s *ProxyService) GetDomains() ([]models.Domain, error) {
	return s.repos.Proxy.GetDomains()
}

// DeleteDomain removes a domain
func (s *ProxyService) DeleteDomain(name string) error {
	if err := s.repos.Proxy.DeleteDomain(name); err != nil {
		return err
	}

	s.logger.Info("Domain deleted", "domain", name)
	return nil
}

// UpdateServiceSubdomain updates the subdomain for a service
func (s *ProxyService) UpdateServiceSubdomain(serviceName, subdomain, domain string) error {
	// Validate subdomain if provided
	if subdomain != "" {
		if err := s.validateSubdomain(subdomain); err != nil {
			return err
		}
	}

	// Get the service
	service, err := s.repos.Service.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("service not found: %w", err)
	}

	// Update service subdomain and domain
	service.Subdomain = strings.ToLower(strings.TrimSpace(subdomain))
	service.Domain = strings.ToLower(strings.TrimSpace(domain))

	if err := s.repos.Service.Update(service); err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	s.logger.Info("Service subdomain updated", "service", serviceName, "subdomain", subdomain, "domain", domain)
	return nil
}

// GetServiceEndpoint returns the full endpoint URL for a service
func (s *ProxyService) GetServiceEndpoint(serviceName string) (string, error) {
	service, err := s.repos.Service.GetByName(serviceName)
	if err != nil {
		return "", fmt.Errorf("service not found: %w", err)
	}

	// Determine subdomain
	subdomain := service.Subdomain
	if subdomain == "" {
		// Auto-generate from service name
		subdomain = service.Name
	}

	// Determine domain
	domain := service.Domain
	if domain == "" {
		// Use primary domain
		primaryDomain, err := s.repos.Proxy.GetPrimaryDomain()
		if err != nil {
			return "", fmt.Errorf("failed to get primary domain: %w", err)
		}
		if primaryDomain == nil {
			return "", fmt.Errorf("no primary domain configured")
		}
		domain = primaryDomain.Name
	}

	// Build endpoint
	endpoint := fmt.Sprintf("%s.%s", subdomain, domain)
	return endpoint, nil
}

// validateDomainName validates a domain name format
func (s *ProxyService) validateDomainName(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain name cannot be empty")
	}

	// Basic domain validation regex
	// Allows: example.com, sub.example.com, localhost, example.local
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}

	return nil
}

// validateSubdomain validates a subdomain format
func (s *ProxyService) validateSubdomain(subdomain string) error {
	if subdomain == "" {
		return fmt.Errorf("subdomain cannot be empty")
	}

	// Subdomain validation: alphanumeric and hyphens, must start with letter/number
	subdomainRegex := regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

	if !subdomainRegex.MatchString(subdomain) {
		return fmt.Errorf("invalid subdomain format: %s (use alphanumeric and hyphens only)", subdomain)
	}

	return nil
}

// GenerateTraefikConfig generates Traefik configuration for enabled services
func (s *ProxyService) GenerateTraefikConfig() (string, error) {
	// Get all enabled services
	services, err := s.repos.Service.GetAll()
	if err != nil {
		return "", fmt.Errorf("failed to get services: %w", err)
	}

	// Get proxy config
	proxyConfig, err := s.repos.Proxy.GetProxyConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get proxy config: %w", err)
	}

	if !proxyConfig.Enabled {
		return "", fmt.Errorf("proxy is not enabled")
	}

	// Build Traefik configuration
	var configBuilder strings.Builder
	configBuilder.WriteString("# Auto-generated Traefik configuration for MediaCheky services\n\n")

	for _, service := range services {
		if !service.Enabled {
			continue
		}

		endpoint, err := s.GetServiceEndpoint(service.Name)
		if err != nil {
			s.logger.Warn("Skipping service without valid endpoint", "service", service.Name, "error", err)
			continue
		}

		// Generate Traefik labels for this service
		configBuilder.WriteString(fmt.Sprintf("# Service: %s\n", service.DisplayName))
		configBuilder.WriteString(fmt.Sprintf("# Endpoint: %s\n", endpoint))
		configBuilder.WriteString(fmt.Sprintf("# Container: %s\n", service.Name))
		configBuilder.WriteString("\n")
	}

	return configBuilder.String(), nil
}

// CheckSubdomainConflict checks if a subdomain is already in use
func (s *ProxyService) CheckSubdomainConflict(subdomain, domain, excludeServiceName string) (bool, string, error) {
	services, err := s.repos.Service.GetAll()
	if err != nil {
		return false, "", fmt.Errorf("failed to get services: %w", err)
	}

	// Normalize input
	subdomain = strings.ToLower(strings.TrimSpace(subdomain))
	domain = strings.ToLower(strings.TrimSpace(domain))

	for _, service := range services {
		if service.Name == excludeServiceName {
			continue
		}

		serviceSubdomain := service.Subdomain
		if serviceSubdomain == "" {
			// Auto-generated subdomain
			serviceSubdomain = service.Name
		}

		serviceDomain := service.Domain
		if serviceDomain == "" {
			// Use primary domain
			primaryDomain, err := s.repos.Proxy.GetPrimaryDomain()
			if err == nil && primaryDomain != nil {
				serviceDomain = primaryDomain.Name
			}
		}

		// Check for conflict
		if strings.ToLower(serviceSubdomain) == subdomain && strings.ToLower(serviceDomain) == domain {
			return true, service.Name, nil
		}
	}

	return false, "", nil
}
