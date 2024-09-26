public sealed partial class MainPage : Page
{
    public MainPage()
    {
        this.InitializeComponent();

        // Événement chargé pour s'assurer que le DataContext est bien prêt
        this.Loaded += MainPage_Loaded;
    }

    private void MainPage_Loaded(object sender, RoutedEventArgs e)
    {
        // Vérifier si le DataContext est bien initialisé
        if (this.DataContext is MainViewModel viewModel)
        {
            // S'assurer que la collection "Applications" est non nulle
            if (viewModel.Applications != null)
            {
                // Écouter les changements dans la collection "Applications"
                viewModel.Applications.CollectionChanged += Applications_CollectionChanged;
            }
        }
    }

    private void Applications_CollectionChanged(object sender, System.Collections.Specialized.NotifyCollectionChangedEventArgs e)
    {
        // Ajouter chaque élément nouvellement ajouté à la Grid
        foreach (string appName in e.NewItems)
        {
            AddControlsDynamically(appName);
        }
    }

    public void AddControlsDynamically(string appName)
    {
        TextBlock textBlock = new TextBlock
        {
            Text = appName,
            FontSize = 18,
            Margin = new Thickness(10)
        };

        MainGrid.Children.Add(textBlock);
    }
}
