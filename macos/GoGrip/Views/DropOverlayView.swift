import SwiftUI

struct DropOverlayView: View {
    var isValidDrop: Bool = true

    var body: some View {
        ZStack {
            Color.black.opacity(0.85)

            RoundedRectangle(cornerRadius: 12)
                .stroke(style: StrokeStyle(lineWidth: 2, dash: [8]))
                .foregroundColor(.accentColor)
                .padding(8)

            VStack(spacing: 12) {
                Image(systemName: "arrow.down.doc")
                    .font(.system(size: 48))
                Text(isValidDrop ? "释放以打开" : "不支持的文件类型")
                    .font(.headline)
            }
            .foregroundColor(.white)
        }
        .transition(.opacity)
        .animation(.easeInOut(duration: 0.2), value: isValidDrop)
    }
}
