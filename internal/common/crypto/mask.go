package crypto

func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

func MaskIDCard(idCard string) string {
	if len(idCard) != 18 {
		return idCard
	}
	return idCard[:6] + "****" + idCard[14:]
}
