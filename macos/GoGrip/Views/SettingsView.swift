import SwiftUI
import ServiceManagement

struct SettingsView: View {
    @State private var launchAtLogin: Bool = false
    @State private var showLoginError = false
    @State private var loginErrorMessage = ""

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("设置").font(.headline)

            Toggle("开机自启", isOn: $launchAtLogin)
                .onChange(of: launchAtLogin) { enabled in
                    toggleLoginItem(enabled: enabled)
                }

            Spacer()
        }
        .padding()
        .frame(width: 260, height: 120)
        .onAppear {
            launchAtLogin = isLoginItemEnabled()
        }
        .alert("Login Item Error", isPresented: $showLoginError) {
            Button("OK", role: .cancel) {}
        } message: {
            Text(loginErrorMessage)
        }
    }

    private func toggleLoginItem(enabled: Bool) {
        let service = SMAppService.mainApp
        do {
            if enabled {
                try service.register()
            } else {
                try service.unregister()
            }
        } catch {
            loginErrorMessage = error.localizedDescription
            showLoginError = true
        }
    }

    private func isLoginItemEnabled() -> Bool {
        SMAppService.mainApp.status == .enabled
    }
}
