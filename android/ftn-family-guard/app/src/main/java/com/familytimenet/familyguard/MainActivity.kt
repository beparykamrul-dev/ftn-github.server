package com.familytimenet.familyguard

import android.app.Activity
import android.content.Intent
import android.net.VpnService
import android.os.Bundle
import android.widget.Button
import android.widget.LinearLayout
import android.widget.TextView
import com.familytimenet.familyguard.dns.FtnDnsVpnService

class MainActivity : Activity() {
    private val vpnRequestCode = 4101

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val status = TextView(this).apply {
            text = "FTN Family Guard\nDNS protection is off"
            textSize = 18f
            setPadding(32, 48, 32, 32)
        }

        val button = Button(this).apply {
            text = "Enable DNS Protection"
            setOnClickListener { requestVpnPermission() }
        }

        setContentView(LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            addView(status)
            addView(button)
        })
    }

    private fun requestVpnPermission() {
        val intent = VpnService.prepare(this)
        if (intent != null) {
            startActivityForResult(intent, vpnRequestCode)
        } else {
            startProtection()
        }
    }

    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)
        if (requestCode == vpnRequestCode && resultCode == RESULT_OK) startProtection()
    }

    private fun startProtection() {
        val intent = Intent(this, FtnDnsVpnService::class.java)
        startService(intent)
    }
}
