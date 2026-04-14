import React from 'react'

export function SettingsPage() {
  return (
    <div style={{ padding: '24px', color: '#fff' }}>
      <h1>Settings & Organization</h1>
      
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '24px', marginTop: '24px' }}>
        <div style={{ backgroundColor: '#1e1e1e', padding: '16px', borderRadius: '8px' }}>
          <h3>Role Based Access Control</h3>
          <p style={{ color: '#aaa', fontSize: '0.9em' }}>Manage users and permissions for your Tenant.</p>
          
          <ul style={{ listStyle: 'none', padding: 0, marginTop: '16px' }}>
            <li style={{ padding: '12px 0', borderBottom: '1px solid #333' }}>
              <strong>Admin User</strong> (admin@gpsgo.com) - <span style={{ color: '#10b981' }}>Super Admin</span>
            </li>
            <li style={{ padding: '12px 0' }}>
              <button style={{ backgroundColor: '#3b82f6', color: '#fff', border: 'none', padding: '8px 16px', borderRadius: '4px', cursor: 'pointer' }}>
                + Invite User
              </button>
            </li>
          </ul>
        </div>
        
        <div style={{ backgroundColor: '#1e1e1e', padding: '16px', borderRadius: '8px' }}>
          <h3>Device Configuration</h3>
          <p style={{ color: '#aaa', fontSize: '0.9em' }}>Global provisioning settings for hardware parsers.</p>
          
          <div style={{ marginTop: '16px' }}>
            <label style={{ display: 'block', marginBottom: '8px' }}>Default Speed Limit (km/h)</label>
            <input type="number" defaultValue={100} style={{ padding: '8px', width: '100px', backgroundColor: '#111', color: '#fff', border: '1px solid #333', borderRadius: '4px' }} />
          </div>
          
          <div style={{ marginTop: '16px' }}>
            <label style={{ display: 'block', marginBottom: '8px' }}>Idle Timeout Threshold (Mins)</label>
            <input type="number" defaultValue={5} style={{ padding: '8px', width: '100px', backgroundColor: '#111', color: '#fff', border: '1px solid #333', borderRadius: '4px' }} />
          </div>
          
          <button style={{ marginTop: '24px', backgroundColor: '#10b981', color: '#fff', border: 'none', padding: '8px 16px', borderRadius: '4px', cursor: 'pointer' }}>
            Save Global Settings
          </button>
        </div>
      </div>
    </div>
  )
}
