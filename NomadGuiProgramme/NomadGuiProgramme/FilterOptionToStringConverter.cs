using System;
using Microsoft.UI.Xaml.Data;


namespace NomadGuiProgramme.Presentation
{
    public class FilterOptionToStringConverter : IValueConverter
    {
        public object Convert(object value, Type targetType, object parameter, string language)
        {
            if (value is FilterOption filterOption)
            {
                return filterOption switch
                {
                    FilterOption.Tous => "Tous",
                    FilterOption.Installés => "Installés",
                    FilterOption.NonInstallés => "Non installés",
                    FilterOption.MiseÀJourDisponible => "Mises à jour disponibles",
                    _ => "Inconnu"
                };
            }
            return "Inconnu";
        }

        public object ConvertBack(object value, Type targetType, object parameter, string language)
        {
            return null;
        }
    }
}
