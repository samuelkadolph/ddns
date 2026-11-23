require "yaml"

require_relative "api"

class Config
  class Error < ArgumentError
  end

  module Validation
    private
    def array?(conf, key, str = key)
      is_a?(conf, key, Array, str)
    end

    def conf?(conf)
      raise Error.new("missing configuration") unless conf.is_a?(Hash)
    end

    def exists?(conf, key, str = key)
      raise Error.new("missing #{str}") unless conf.key?(key)
    end

    def exists_and_array?(conf, key, str = key)
      exists?(conf, key, str)
      array?(conf, key, str)
    end

    def exists_and_hash?(conf, key, str = key)
      exists?(conf, key, str)
      hash?(conf, key, str)
    end

    def exists_and_string?(conf, key, str = key)
      exists?(conf, key, str)
      string?(conf, key, str)
    end

    def hash?(conf, key, str = key)
      is_a?(conf, key, Hash, str)
    end

    def is_a?(conf, key, klass, str = key)
      raise Error.new("#{str} must be a #{klass}") unless conf[key].is_a?(klass)
    end

    def string?(conf, key, str = key)
      is_a?(conf, key, String, str)
    end
  end

  include Validation

  attr_reader :credentials, :domains, :users

  def initialize(path)
    conf = YAML.load_file(path, aliases: true)

    exists_and_hash?(conf, "credentials")
    exists_and_hash?(conf, "domains")
    exists_and_array?(conf, "users")

    @credentials = conf["credentials"].transform_values do |c|
      raise Error.new("credential missing key type") unless c.key?("type")
      raise Error.new("unknown credential type '#{c["type"]}'") unless CREDENTIAL_TYPES.key?(c["type"])
      CREDENTIAL_TYPES[c["type"]].new(c)
    end

    @domains = conf["domains"].transform_values { |c| Domain.new(c) }

    @users = conf["users"].map { |c| User.new(c) }

    domains.each do |n, d|
      raise Error.new("domain '#{n}' uses nonexsitent credentials '#{d.credentials}'") unless credentials.key?(d.credentials)

      d.users.each do |user|
        raise Error.new("domain '#{n}' uses nonexsitent user '#{user}'") unless users.any? { |u| user == u.username }
      end
    end
  end

  class Credential
    include Validation

    def initialize(conf)
      raise NoMethodError
    end

    def to_api
      raise NoMethodError
    end
  end
  CREDENTIAL_TYPES = {}

  class OVH < Credential
    attr_reader :app_key, :app_secret, :consumer_key, :endpoint

    def initialize(conf)
      conf?(conf)
      exists_and_string?(conf, "app_key")
      exists_and_string?(conf, "app_secret")
      exists_and_string?(conf, "consumer_key")
      exists_and_string?(conf, "endpoint")

      @app_key = conf["app_key"]
      @app_secret = conf["app_secret"]
      @consumer_key = conf["consumer_key"]
      @endpoint = conf["endpoint"]
    end

    def to_api
      API::OVH.new(endpoint, app_key, app_secret, consumer_key)
    end
  end
  CREDENTIAL_TYPES["ovh"] = OVH

  class Domain
    include Validation

    attr_reader :credentials, :users, :zone

    def initialize(conf)
      conf?(conf)
      exists_and_string?(conf, "credentials")
      exists?(conf, "users")
      exists_and_string?(conf, "zone")

      @credentials = conf["credentials"]
      @users = Array(conf["users"])
      @zone = conf["zone"]
    end
  end

  class User
    include Validation

    attr_reader :password, :username

    def initialize(conf)
      conf?(conf)
      exists_and_string?(conf, "password")
      exists_and_string?(conf, "username")

      @password = conf["password"]
      @username = conf["username"]
    end
  end
end
