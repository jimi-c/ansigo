package inventory

import ()

type InventoryManager struct {

}

func (im *InventoryManager) GetHosts() []Host {
  return []Host{
    *NewHost("host01", nil),
    *NewHost("host02", nil),
    *NewHost("host03", nil),
    *NewHost("host04", nil),
    *NewHost("host05", nil),
    *NewHost("host06", nil),
    *NewHost("host07", nil),
    *NewHost("host08", nil),
    *NewHost("host09", nil),
    *NewHost("host10", nil),
  }
}
