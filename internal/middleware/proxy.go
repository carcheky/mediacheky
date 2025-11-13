package middleware

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// ReverseProxy creates a middleware that proxies requests based on Host header
func ReverseProxy(repos *repository.Repositories, log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Host header
		host := c.Hostname()
		if host == "" {
			// No Host header, continue to normal routing
			return c.Next()
		}

		log.Debug("Proxy middleware checking host", "host", host)

		// Proxy is always enabled - no need to check database config
		// Parse subdomain and domain from host
		parts := strings.Split(host, ".")
		if len(parts) < 2 {
			// Not a subdomain, continue to normal routing
			log.Debug("Not a subdomain, skipping", "host", host, "parts", len(parts))
			return c.Next()
		}

		subdomain := parts[0]
		domain := strings.Join(parts[1:], ".")

		log.Debug("Parsed host", "subdomain", subdomain, "domain", domain)

		// Check if domain is registered
		domains, err := repos.Proxy.GetDomains()
		if err != nil {
			log.Error("Failed to get domains", "error", err)
			return c.Next()
		}

		domainExists := false
		for _, d := range domains {
			if strings.EqualFold(d.Name, domain) {
				domainExists = true
				break
			}
		}

		if !domainExists {
			log.Debug("Domain not registered", "domain", domain)
			return c.Next()
		}

		// Find service with matching subdomain and domain
		services, err := repos.Service.GetAll()
		if err != nil {
			log.Error("Failed to get services", "error", err)
			return c.Next()
		}

		var targetService *struct {
			Name          string
			URL           string
			Port          int
			Subdomain     string
			Domain        string
			Enabled       bool
			ContainerName string
		}

		for _, service := range services {
			if !service.Enabled {
				continue
			}

			// Determine service subdomain (use service name if not set)
			serviceSubdomain := service.Subdomain
			if serviceSubdomain == "" {
				serviceSubdomain = service.Name
			}

			// Determine service domain (use primary if not set, fallback to docker.internal)
			serviceDomain := service.Domain
			if serviceDomain == "" {
				primaryDomain, err := repos.Proxy.GetPrimaryDomain()
				if err == nil && primaryDomain != nil {
					serviceDomain = primaryDomain.Name
				} else {
					// Fallback to docker.internal if no primary domain configured
					serviceDomain = "docker.internal"
					log.Debug("Using fallback domain", "domain", serviceDomain)
				}
			}

			// Check if this service matches the requested subdomain + domain
			if strings.EqualFold(serviceSubdomain, subdomain) && strings.EqualFold(serviceDomain, domain) {
				// Get port from service (use default if not set)
				port := service.Port
				if port == 0 {
					// Default ports for common services
					switch service.Name {
					case "radarr":
						port = 7878
					case "sonarr":
						port = 8989
					case "jellyfin":
						port = 8096
					case "prowlarr":
						port = 9696
					case "bazarr":
						port = 6767
					case "qbittorrent":
						port = 8080
					case "jellyseerr":
						port = 5055
					case "jellystat":
						port = 3000
					default:
						port = 8080
					}
				}

				// Build container URL using actual container hostname
				// Try multiple hostname options:
				// 1. Container ID if available (most reliable)
				// 2. Service name (default Docker Compose behavior)
				containerHost := service.Name
				if service.ContainerID != "" {
					// Use container ID as hostname (Docker allows this)
					containerHost = service.ContainerID[:12] // Use short ID
				}

				targetService = &struct {
					Name          string
					URL           string
					Port          int
					Subdomain     string
					Domain        string
					Enabled       bool
					ContainerName string
				}{
					Name:          service.Name,
					URL:           fmt.Sprintf("http://%s:%d", containerHost, port),
					Port:          port,
					Subdomain:     serviceSubdomain,
					Domain:        serviceDomain,
					Enabled:       service.Enabled,
					ContainerName: containerHost,
				}

				log.Debug("Found matching service",
					"service", service.Name,
					"subdomain", serviceSubdomain,
					"domain", serviceDomain,
					"containerHost", containerHost,
					"port", port,
					"targetURL", targetService.URL,
				)

				break
			}
		}

		if targetService == nil {
			log.Debug("No matching service found", "subdomain", subdomain, "domain", domain)
			return c.Next()
		}

		log.Info("Proxying request",
			"host", host,
			"service", targetService.Name,
			"target", targetService.URL,
			"path", c.Path(),
		)

		// Parse target URL
		targetURL, err := url.Parse(targetService.URL)
		if err != nil {
			log.Error("Invalid target URL", "url", targetService.URL, "error", err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "Invalid proxy target configuration",
			})
		}

		// Preserve the original path and query
		targetURL.Path = c.Path()
		targetURL.RawQuery = string(c.Request().URI().QueryString())

		// Set proxy headers to preserve original host and configure URL base
		c.Request().Header.Set("X-Forwarded-Host", host)
		c.Request().Header.Set("X-Forwarded-Proto", "http")
		c.Request().Header.Set("X-Real-IP", c.IP())
		c.Request().Header.Set("X-Forwarded-For", c.IP())

		// Set Host header to target service (required for proper routing)
		c.Request().Header.SetHost(targetURL.Host)

		// Use Fiber's built-in proxy middleware (no redirect following to avoid localhost issues)
		if err := proxy.Do(c, targetURL.String()); err != nil {
			log.Error("Proxy request failed",
				"target", targetURL.String(),
				"service", targetService.Name,
				"containerHost", targetService.ContainerName,
				"port", targetService.Port,
				"error", err,
			)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":         "Failed to proxy request to service",
				"service":       targetService.Name,
				"containerHost": targetService.ContainerName,
				"targetURL":     targetURL.String(),
				"details":       err.Error(),
			})
		}

		// Modify response
		c.Response().Header.Del(fiber.HeaderServer)

		return nil
	}
}
