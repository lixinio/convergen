// Code generated by github.com/reedom/convergen
// DO NOT EDIT.

package multi_intf

type DomainModel struct {
	ID string
}

type TransportModel struct {
	ID string
}

type StorageModel struct {
	ID string
}

func (d *DomainModel) ToStorage() (dst *StorageModel) {
	dst = &StorageModel{}
	dst.ID = d.ID

	return
}

func (d *DomainModel) ToTransport() (dst *TransportModel) {
	dst = &TransportModel{}
	dst.ID = d.ID

	return
}

func (s *StorageModel) ToDomain() (dst *DomainModel) {
	dst = &DomainModel{}
	dst.ID = s.ID

	return
}

func (s *StorageModel) ToTransport() (dst *TransportModel) {
	dst = &TransportModel{}
	dst.ID = s.ID

	return
}