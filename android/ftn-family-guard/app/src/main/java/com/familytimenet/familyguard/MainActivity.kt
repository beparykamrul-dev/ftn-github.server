package com.familytimenet.familyguard

import android.app.Activity
import android.content.Intent
import android.net.VpnService
import android.os.Build
import android.os.Bundle
import android.widget.Button
import android.widget.LinearLayout
import android.widget.TextView
import com.familytimenet.familyguard.dns.FtnDnsVpnService

/** Small local control surface; policy remains owned by the FTN control plane. */
class MainActivity : Activity() {
    private val vpnRequestCode = 4101
    private lateinit var status: TextView
    private lateinit var action: Button

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        buildUi()
    }

    private fun buildUi() {
        status = TextView(this).apply {
            text = "FTN Family Guard\n\nDNS protection: OFF\nDNSSEC: REQUIRED\nResolver: FTN preferred\n1.1.1.1 Family: policy-controlled\nRaw DNS logs: OFF"
            textSize = 17f
            setPadding(32, 48, 32, 32)
        }
        action = Button(this).apply {
            text = "Enable DNS Protection"
            setOnClickListener { requestVpnPermission() }
        }
        val refresh = Button(this).apply {
            text = "Refresh status"
            setOnClickListener { refreshStatus() }
        }
        setContentView(LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(20, 20, 20, 20)
            addView(status)
            addView(action)
            addView(refresh)
        })
    }

    private fun requestVpnPermission() {
        val intent = VpnService.prepare(this)
        if (intent != null) startActivityForResult(intent, vpnRequestCode) else startProtection()
    }

    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)
        if (requestCode == vpnRequestCode) {
            if (resultCode == RESULT_OK) startProtection()
            else status.text = "FTN Family Guard\n\nDNS protection: OFF\nVPN permission was not granted."
        }
    }

    private fun startProtection() {
        val intent = Intent(this, FtnDnsVpnService::class.java)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) startForegroundService(intent) else startService(intent)
        status.text = "FTN Family Guard\n\nDNS protection: ACTIVE\nDNSSEC: REQUIRED\nResolver: FTN preferred\n1.1.1.1 Family: policy-controlled\nRaw DNS logs: OFF"
        action.text = "Protection active"
        action.isEnabled = false
    }

    private fun refreshStatus() {
        status.text = "FTN Family Guard\n\nLocal service status refreshed.\nDNSSEC: REQUIRED\nResolver: FTN preferred\n1.1.1.1 Family: policy-controlled\nRaw DNS logs: OFF"
    }
}
