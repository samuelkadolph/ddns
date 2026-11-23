require "bundler/setup"
require "sinatra/base"

Bundler.require(:default, Sinatra::Base.environment)

require "sinatra/reloader" if Sinatra::Base.development?

require "base64"
require "ipaddr"
require "openssl"

require_relative "config"

class App < Sinatra::Application
  configure do
    set :config, Config.new("/config/ddns.yaml")
  end

  configure :development do
    register Sinatra::Reloader
    also_reload "lib/*.rb"
  end

  get "/nic/update" do
    return plain_text(401, "badauth") unless user = find_user
    return plain_text(400, "badagent") unless params.key?(:hostname)
    return plain_text(400, "badagent") unless params.key?(:myip)

    begin
      ip = IPAddr.new(params[:myip])
    rescue IPAddr::InvalidAddressError
      return plain_text(400, "badagent")
    end

    host = params[:hostname]

    return plain_text(400, "nohost") unless domain = find_domain(host)
    return plain_text(400, "nohost") unless domain.users.include?(user.username)

    status, msg = find_credentials(domain).to_api.update(domain.zone, host, ip)

    if status == :no_change
      plain_text(200, "nochg " << msg)
    elsif status == :error
      plain_text(500, "911 " << msg.to_s)
    elsif status == :success
      plain_text(200, "good " << msg)
    end
  end

  private
  def config
    settings.config
  end

  def find_credentials(domain)
    config.credentials[domain.credentials]
  end

  def find_domain(hostname)
    config.domains[hostname]
  end

  def find_user
    return unless auth = request.env["HTTP_AUTHORIZATION"]

    scheme, credentials = auth.split(" ", 2)
    return unless scheme == "Basic"

    decoded = Base64.decode64(credentials)
    user, pass = decoded.split(":", 2)

    config.users.find { |u| u.username == user && OpenSSL.fixed_length_secure_compare(u.password, pass) }
  end

  def plain_text(code, text)
    [ code, { "Content-Type" => "text/plain" }, text + "\n" ]
  end
end
