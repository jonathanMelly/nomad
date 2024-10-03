using System.Collections.ObjectModel;

namespace NomadGuiProgramme.Presentation
{
    public sealed partial class MainPage : Page
    {
        public ObservableCollection<ApplicationItem> Applications { get; } = new ObservableCollection<ApplicationItem>();

        public MainViewModel ViewModel { get; } = new MainViewModel();


        public MainPage()
        {
            this.InitializeComponent();
            this.DataContext = new MainViewModel();
        }

    }
    public class ApplicationItem
{
    public string Name { get; set; }
    public ICommand InstallCommand { get; }

    public ApplicationItem(string name, ICommand installAppCommand)
    {
        Name = name;
        InstallCommand = new RelayCommand(() => installAppCommand.Execute(name));
    }
}
}
