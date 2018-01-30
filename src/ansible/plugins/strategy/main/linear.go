package main

import (
  strategy_base "ansible/plugins/strategy"
)

type StrategyPlugin struct {
  strategy_base.StrategyPluginBase
}

var Strategy StrategyPlugin
