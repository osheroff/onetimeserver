#!/usr/bin/env ruby

WRAPPER = File.dirname(__FILE__) + "/../onetimeserver"
require 'json'

mysql = JSON.parse(`#{WRAPPER}`)
puts "Booted a mysql at port %d in path %s." % [mysql['port'], mysql['mysql_path']]
puts "It will exit when PID %d exits (this script)" % [mysql['parent_pid']]
sleep


