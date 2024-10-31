

using System.Reflection;
using Microsoft.UI.Xaml.Media;
using Microsoft.UI.Xaml.Media.Imaging;

namespace NomadGuiProgramme.Presentation
{
    public sealed partial class MainPage : Page
    {
        public MainViewModel ViewModel { get; } = new MainViewModel();

        public MainPage()
        {
            this.InitializeComponent();
            string relativePath = "..\\..\\..\\..\\image\\R.png"; // Sans "../"
            string basePath = Path.GetDirectoryName(Assembly.GetExecutingAssembly().Location);
            string absolutePath = Path.Combine(basePath, relativePath);
            ImageSource imageSource = new BitmapImage(new Uri(absolutePath));
            myImage.Source = imageSource;
        }
    }
}
