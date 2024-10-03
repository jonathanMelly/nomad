

namespace NomadGuiProgramme.Presentation
{
    public sealed partial class MainPage : Page
    {
        public MainViewModel ViewModel { get; } = new MainViewModel();

        public MainPage()
        {
            this.InitializeComponent();
        }
    }
}
