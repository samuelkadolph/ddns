require "digest"
require "json"
require "net/http"
require "uri"

class API
  def initialize(*)
    raise NoMethodError
  end

  def update(zone, domain, ip)
    raise NoMethodError
  end

  class OVH < API
    ENDPOINTS = {
      "ovh-ca" => "https://ca.api.ovh.com/1.0/",
      "ovh-eu" => "https://eu.api.ovh.com/1.0/",
      "ovh-us" => "https://api.us.ovhcloud.com/1.0/",
    }

    def initialize(endpoint, app_key, app_secret, consumer_key)
      raise ArgumentError.new("unknown endpoint #{endpoint}") unless ENDPOINTS.key?(endpoint)

      @api = ENDPOINTS[endpoint]
      @app_key = app_key
      @app_secret = app_secret
      @consumer_key = consumer_key
    end

    def update(zone, domain, ip)
      zone_uri = URI(@api) + "domain/zone/#{zone}/"
      type = ip.ipv6? ? "AAAA" : "A"
      subdomain = domain.delete_suffix(zone).chomp(".")

      Net::HTTP.start(zone_uri.hostname, use_ssl: true) do |h|
        records_uri = zone_uri.dup + "record/"
        records_uri.query = URI.encode_www_form(fieldType: type, subDomain: subdomain)
        resp = h.request(sign_request(Net::HTTP::Get.new(records_uri)))
        return [:error, "records error"] unless resp.code == "200"
        records = JSON.parse(resp.body)
        return [:error, "no record found"] if records.empty?

        record_uri = records_uri.dup + records[0].to_s
        resp = h.request(sign_request(Net::HTTP::Get.new(record_uri)))
        return [:error, "record error"] unless resp.code == "200"
        record = JSON.parse(resp.body)

        if ip == record["target"]
          return [:no_change, ip.to_s]
        else
          req = Net::HTTP::Put.new(record_uri)
          req.body = { target: ip.to_s }.to_json
          req.content_type = "application/json"
          resp = h.request(sign_request(req))
          return [:error, "update error"] unless resp.code == "200"
        end

        refresh_uri = zone_uri.dup + "refresh/"
        resp = h.request(sign_request(Net::HTTP::Post.new(refresh_uri)))
        return [:error, "refresh error"] unless resp.code == "200"

        puts "Updated #{domain} to #{ip}"

        [:success, ip.to_s]
      end
    end

    private
    def generate_signature(method, uri, body, time)
      "$1$" << Digest::SHA1.hexdigest([@app_secret, @consumer_key, method, uri, body, time].join("+"))
    end

    def sign_request(request)
      request["X-Ovh-Application"] = @app_key
      request["X-Ovh-Consumer"] = @consumer_key
      request["X-Ovh-Signature"] = generate_signature(request.method, request.uri, request.body, time = Time.now.to_i)
      request["X-Ovh-Timestamp"] = time
      request
    end
  end
end
