package route53

import (
	"fmt"
	"time"

	"github.com/samuelkadolph/ddns/internal/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Manager struct {
	config     *config.Config
	serviceMap map[string]*route53.Route53
}

func NewManager(cfg *config.Config) (*Manager, error) {
	manager := &Manager{config: cfg, serviceMap: make(map[string]*route53.Route53)}

	return manager, nil
}

func (m *Manager) UpdateDomain(domain *config.Domain, ip string) error {
	r53 := m.route53ForDomain(domain)

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  buildChangeBatch(domain.Name, ip, m.config.TTL),
		HostedZoneId: aws.String(fmt.Sprintf("/hostedzone/%s", domain.ZoneID)),
	}
	if _, err := r53.ChangeResourceRecordSets(input); err != nil {
		return err
	}

	return nil
}

func (m *Manager) route53ForCredentials(creds *config.Credentials) *route53.Route53 {
	if r53, ok := m.serviceMap[creds.Name]; ok {
		return r53
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(creds.AccessID, creds.AccessKey, ""),
	}))

	m.serviceMap[creds.Name] = route53.New(sess)

	return m.serviceMap[creds.Name]
}

func (m *Manager) route53ForDomain(domain *config.Domain) *route53.Route53 {
	return m.route53ForCredentials(m.config.Credentials[domain.Credentials])
}

func buildChangeBatch(domain string, ip string, ttl int) *route53.ChangeBatch {
	return &route53.ChangeBatch{
		Changes: []*route53.Change{
			{
				Action: aws.String("UPSERT"),
				ResourceRecordSet: &route53.ResourceRecordSet{
					Name: aws.String(domain),
					ResourceRecords: []*route53.ResourceRecord{
						{
							Value: aws.String(ip),
						},
					},
					TTL:  aws.Int64(int64(ttl)),
					Type: aws.String("A"),
				},
			},
		},
		Comment: aws.String(fmt.Sprintf("ddns update %s", time.Now().Format("2006-01-02T15:04:05-0700"))),
	}
}
