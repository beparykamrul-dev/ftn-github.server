package com.familytimenet.familyguard.dns

/** Parses only the DNS header and first question name/type/class. */
object DnsQuestionParser {
    private const val HEADER_SIZE = 12
    private const val MAX_NAME_LENGTH = 253
    private const val MAX_LABEL_LENGTH = 63

    fun parse(packet: ByteArray): DnsQueryMetadata? {
        if (packet.size < HEADER_SIZE) return null
        val flags = u16(packet, 2)
        val questions = u16(packet, 4)
        if ((flags and 0x8000) != 0 || questions != 1) return null

        var offset = HEADER_SIZE
        val labels = ArrayList<String>()
        var totalLength = 0
        while (offset < packet.size) {
            val length = packet[offset].toInt() and 0xff
            offset++
            if (length == 0) break
            if (length > MAX_LABEL_LENGTH || offset + length > packet.size) return null
            totalLength += length + 1
            if (totalLength > MAX_NAME_LENGTH + 1) return null
            labels += packet.copyOfRange(offset, offset + length).toString(Charsets.US_ASCII)
            offset += length
        }
        if (labels.isEmpty() || offset + 4 > packet.size) return null

        val domain = labels.joinToString(".").lowercase()
        if (!isValidDomain(domain)) return null

        return DnsQueryMetadata(
            transactionId = u16(packet, 0),
            domain = domain,
            recordType = u16(packet, offset),
            recordClass = u16(packet, offset + 2)
        )
    }

    private fun isValidDomain(domain: String): Boolean =
        domain.length <= MAX_NAME_LENGTH &&
            domain.split('.').all { it.isNotEmpty() && it.length <= MAX_LABEL_LENGTH &&
                it.all { c -> c.isLetterOrDigit() || c == '-' || c == '_' }
            }

    private fun u16(bytes: ByteArray, offset: Int): Int =
        ((bytes[offset].toInt() and 0xff) shl 8) or (bytes[offset + 1].toInt() and 0xff)
}
