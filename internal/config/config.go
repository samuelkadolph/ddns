package config

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Credentials map[string]*Credentials `yaml:"credentials"`
	Domains     map[string]*Domain      `yaml:"domains"`
	TTL         int                     `yaml:"ttl"`
	Users       []*User                 `yaml:"users"`
}

type Credentials struct {
	AccessID  string `yaml:"access_id"`
	AccessKey string `yaml:"access_key"`
	Name      string
}

type Domain struct {
	Credentials string `yaml:"credentials"`
	Name        string
	Users       []string `yaml:"users"`
	ZoneID      string   `yaml:"zone_id"`
}

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func Load(data []byte) (*Config, error) {
	c := &Config{}

	err := yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	if err = c.validate(); err != nil {
		return nil, err
	}

	for name, c := range c.Credentials {
		c.Name = name
	}

	for name, d := range c.Domains {
		d.Name = name
	}

	if c.TTL == 0 {
		c.TTL = 60
	}

	return c, nil
}

func Read(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Load(data)
}
func (c *Config) AuthenticateUser(username string, password string) (*User, bool) {
	user, ok := c.FindUser(username)
	if !ok {
		return nil, false
	}

	if subtle.ConstantTimeEq(int32(len(user.Password)), int32(len(password))) != 1 ||
		subtle.ConstantTimeCompare([]byte(password), []byte(user.Password)) != 1 {
		return nil, false
	}

	return user, true
}

func (c *Config) AuthorizeUser(user *User, domain *Domain) bool {
	for _, allowed_user := range domain.Users {
		if user.Username == allowed_user {
			return true
		}
	}

	return false
}

func (c *Config) FindUser(username string) (*User, bool) {
	for _, user := range c.Users {
		if user.Username == username {
			return user, true
		}
	}

	return nil, false
}

func (c *Config) validate() error {
	if len(c.Credentials) == 0 {
		return errors.New("must specify at least one set of credentials")
	}

	for name, creds := range c.Credentials {
		if creds.AccessID == "" {
			return errors.New(fmt.Sprintf("credentials '%s' must specify access_id", name))
		}

		if creds.AccessKey == "" {
			return errors.New(fmt.Sprintf("credentials '%s' must specify access_key", name))
		}
	}

	if len(c.Users) == 0 {
		return errors.New("must specify at least one user")
	}

	for _, user := range c.Users {
		if user.Username == "" {
			return errors.New(fmt.Sprintf("all users must specify a username"))
		}

		if user.Password == "" {
			return errors.New(fmt.Sprintf("user '%s' must specify a password", user.Username))
		}
	}

	for name, domain := range c.Domains {
		if _, ok := c.Credentials[domain.Credentials]; !ok {
			return errors.New(fmt.Sprintf("domain '%s': could not find credentials named '%s'", name, domain.Credentials))
		}

		for _, user := range domain.Users {
			found := false

			for _, existing_user := range c.Users {
				if user == existing_user.Username {
					found = true
					break
				}
			}

			if !found {
				return errors.New(fmt.Sprintf("domain '%s': could not find user named '%s'", name, user))
			}
		}

		if domain.ZoneID == "" {
			return errors.New(fmt.Sprintf("domain '%s' must specify zone_id", name))
		}
	}

	return nil
}
