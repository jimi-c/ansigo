package inventory

import ()

type InventoryManager struct {

}

func (im *InventoryManager) GetHosts() []Host {
  return []Host{
    Host{"host01"},
    Host{"host02"},
    Host{"host03"},
    Host{"host04"},
    Host{"host05"},
    Host{"host06"},
    Host{"host07"},
    Host{"host08"},
    Host{"host09"},
    Host{"host10"},
  }
}
