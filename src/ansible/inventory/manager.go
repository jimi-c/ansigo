package inventory

import ()

type InventoryManager struct {

}

func (im *InventoryManager) GetHosts() []Host {
  return []Host{
    *NewHost("192.168.122.100", nil),
  }
}
